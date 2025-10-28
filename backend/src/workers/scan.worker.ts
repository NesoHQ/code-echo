import { Worker, Job } from 'bullmq';
import path from 'path';
import fs from 'fs';
import simpleGit from 'simple-git';
import unzipper from 'unzipper';
import { v4 as uuidv4 } from 'uuid';
import { connection, closeQueue } from '../services/queue.service';
import { runCodeEchoCLI } from '../services/cli-runner';
import { updateJobStatus } from '../services/job.service';

const QUEUE_NAME = 'scan-jobs';
const CONCURRENCY = Number(process.env.WORKER_CONCURRENCY ?? 2);
const PREFIX = process.env.BULLMQ_PREFIX ?? 'codeecho';

const MAX_EXTRACT_FILES = Number(process.env.MAX_EXTRACT_FILES ?? 5000);
const MAX_ZIP_SIZE = Number(process.env.MAX_ZIP_SIZE_BYTES ?? 200 * 1024 * 1024);
const GIT_CLONE_TIMEOUT_MS = Number(process.env.GIT_CLONE_TIMEOUT_MS ?? 120_000);
const SCAN_TIMEOUT_MS = Number(process.env.SCAN_TIMEOUT_MS ?? 10 * 60_000);

console.log('scan.worker starting — QUEUE_NAME=', QUEUE_NAME, ' TEMP_DIR=', process.env.TEMP_DIR);

interface ScanJobData {
	jobId?: string;
	repoUrl?: string;
	zipPath?: string;
	source?: 'git' | 'upload' | 'path';
}

// safety: prevent zip-slip
function safeResolveExtract(dest: string, entryPath: string) {
	const resolved = path.resolve(dest, entryPath);
	if (!resolved.startsWith(dest + path.sep)) {
		throw new Error('Zip contains invalid entry (zip-slip)');
	}
	return resolved;
}

async function extractZipSafe(zipPath: string, dest: string) {
	const st = await fs.promises.stat(zipPath);
	if (st.size > MAX_ZIP_SIZE) throw new Error('Zip too large');

	const dir = await unzipper.Open.file(zipPath);
	if (dir.files.length > MAX_EXTRACT_FILES) throw new Error('Zip has too many files');

	for (const entry of dir.files) {
		const target = safeResolveExtract(dest, entry.path);
		if (entry.type === 'Directory') {
			await fs.promises.mkdir(target, { recursive: true });
		} else {
			await fs.promises.mkdir(path.dirname(target), { recursive: true });
			await new Promise<void>((resolve, reject) => {
				entry
					.stream()
					.pipe(fs.createWriteStream(target, { mode: 0o600 }))
					.on('finish', () => resolve())
					.on('error', (e) => reject(e));
			});
		}
	}
}

async function cloneRepo(repoUrl: string, dest: string) {
	try {
		const exists = await fs.promises
			.stat(dest)
			.then(() => true)
			.catch(() => false);

		if (exists) {
			const files = await fs.promises.readdir(dest);
			if (files.length > 0) {
				console.log(`[worker] Cleaning existing directory before clone: ${dest}`);
				await fs.promises.rm(dest, { recursive: true, force: true });
			}
		}
	} catch (err) {
		console.warn(`[worker] Failed to check or clean existing directory ${dest}:`, err);
	}
	const git = simpleGit();
	await git.clone(repoUrl, dest, ['--depth', '1']);
}

async function runWithTimeout<T>(fn: () => Promise<T>, ms: number, onTimeout?: () => Promise<void>) {
	let timedOut = false;
	const timer = new Promise<never>((_, rej) => {
		const t = setTimeout(() => {
			timedOut = true;
			(onTimeout ? onTimeout() : Promise.resolve()).finally(() => rej(new Error('operation timed out')));
		}, ms);
	});
	try {
		return await Promise.race([fn(), timer]);
	} finally {
		if (timedOut) {
			// nothing
		}
	}
}

async function processScanJob(job: Job<ScanJobData>) {
	// Use DB-safe UUID as our job primary key
	const dbJobId = job.data.jobId ?? uuidv4();
	console.log(`Worker started job dbJobId=${dbJobId} (queue id=${String(job.id)})`);

	const root = path.resolve(process.env.TEMP_DIR ?? '/app/data/outputs', dbJobId);
	await fs.promises.mkdir(root, { recursive: true });

	const logPath = path.join(root, 'run.log');
	const logStream = fs.createWriteStream(logPath, { flags: 'a', mode: 0o600 });
	const log = (...args: any[]) => {
		const line = `[${new Date().toISOString()}] ` + args.map(String).join(' ');
		console.log(line);
		logStream.write(line + '\n');
	};
	log(`Workspace root resolved to: ${root}`);

	try {
		await updateJobStatus(dbJobId, 'running', { progress: '5' });

		if (job.data.repoUrl) {
			log(`Cloning ${job.data.repoUrl} -> ${root}`);
			await runWithTimeout(() => cloneRepo(job.data.repoUrl!, root), GIT_CLONE_TIMEOUT_MS);
		} else if (job.data.zipPath) {
			log(`Extracting zip ${job.data.zipPath} -> ${root}`);
			await extractZipSafe(job.data.zipPath, root);
		} else {
			log(`No repoUrl/zipPath provided; scanning existing workspace ${root}`);
		}

		log(`Running CodeEcho CLI for job ${dbJobId}`);
		await updateJobStatus(dbJobId, 'running', { progress: '25' });

		await runWithTimeout(
			() => runCodeEchoCLI({ workspace: root, jobId: dbJobId, flags: [], logStream }),
			SCAN_TIMEOUT_MS,
			async () => {
				log('Scan timed out');
			}
		);

		const outputPath = path.join(root, 'output.xml');
		await fs.promises.access(outputPath);

		await updateJobStatus(dbJobId, 'done', { resultUrl: outputPath, progress: '100' });
		log(`Job ${dbJobId} done -> ${outputPath}`);
		logStream.end();
		return { status: 'done', output: outputPath };
	} catch (err) {
		const msg = err instanceof Error ? err.message : String(err);
		console.error(`Job ${dbJobId} failed:`, msg);
		log(`ERROR: ${msg}`);
		try {
			await updateJobStatus(dbJobId, 'failed', { error: msg, progress: '100' });
		} catch (e) {
			console.error('Failed to update DB:', e);
		}
		logStream.end();
		throw new Error(msg);
	}
}

// Create worker
export const scanWorker = new Worker<ScanJobData>(QUEUE_NAME, processScanJob, {
	connection,
	concurrency: CONCURRENCY,
	prefix: PREFIX,
});

scanWorker.on('completed', (jobId, returnValue) => {
	console.log(`Job ${String(jobId)} completed. result:`, returnValue);
});
scanWorker.on('failed', (jobId, failedReason) => {
	console.error(`Job ${String(jobId)} failed:`, failedReason);
});
scanWorker.on('error', (err) => {
	console.error('Worker error:', err);
});

// graceful shutdown
async function shutdown(signal: string) {
	console.log(`Received ${signal}, shutting down worker`);
	try {
		await scanWorker.close();
		await closeQueue();
		console.log('Graceful shutdown complete');
		process.exit(0);
	} catch (err) {
		console.error('Error during shutdown:', err);
		process.exit(1);
	}
}

process.on('SIGINT', () => shutdown('SIGINT'));
process.on('SIGTERM', () => shutdown('SIGTERM'));
