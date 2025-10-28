import { spawn } from 'child_process';
import fs from 'fs';
import path from 'path';

export interface RunOpts {
	workspace: string;
	jobId: string;
	flags?: string[];
	logStream?: fs.WriteStream;
}

export function runCodeEchoCLI(opts: RunOpts): Promise<void> {
	const { workspace, jobId, flags = [], logStream } = opts;
	return new Promise((resolve, reject) => {
		const bin = process.env.CODEECHO_PATH ?? '/usr/local/bin/codeecho';
		console.log(`Using CodeEcho binary: ${bin}`);
		const args = ['scan', workspace, '--format', 'xml', '-o', path.join(workspace, 'output.xml'), ...flags];

		const child = spawn(bin, args, {
			stdio: ['ignore', 'pipe', 'pipe'],
			cwd: workspace,
			env: { ...process.env },
		});

		child.stdout.on('data', (d) => {
			const s = d.toString();
			(logStream ?? process.stdout).write(`[${jobId}] ${s}`);
		});

		child.stderr.on('data', (d) => {
			const s = d.toString();
			(logStream ?? process.stderr).write(`[${jobId}][ERR] ${s}`);
		});

		child.on('error', (err) => reject(err));
		child.on('close', (code, signal) => {
			if (code === 0) resolve();
			else reject(new Error(`codeecho exited code=${code} signal=${signal}`));
		});
	});
}
