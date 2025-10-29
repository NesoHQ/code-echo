import { Request, Response } from 'express';
import path from 'path';
import fs from 'fs';
import { v4 as uuidv4 } from 'uuid';
import { createJob } from '../services/job.service';
import { sendToQueue } from '../services/queue.service';

export const ScanController = {
	async startScan(req: Request, res: Response) {
		try {
			const repoUrl = (req.body.repoUrl || req.query.repoUrl) as string | undefined;
			const localPath = (req.body.path || req.query.path) as string | undefined;
			const file = req.file;

			if (!repoUrl && !file && !localPath) {
				return res.status(400).json({ error: 'Provide repoUrl or upload a zip file or a local path' });
			}

			const jobId = uuidv4();
			const tempDir = process.env.TEMP_DIR ?? '/app/data/outputs';
			const workspacePath = path.join(tempDir, jobId);

			// Ensure workspace exists
			await fs.promises.mkdir(workspacePath, { recursive: true });

			// If a zip was uploaded, move it into workspace
			let zipPath: string | undefined;
			if (file) {
				const dest = path.join(workspacePath, 'upload.zip');
				await fs.promises.rename(file.path, dest);
				zipPath = dest;
			}

			// Persist to DB
			await createJob({
				id: jobId,
				source: repoUrl ? 'git' : file ? 'upload' : 'path',
				repoUrl,
				zipPath,
				workspacePath,
			});

			// Enqueue worker job
			await sendToQueue(
				'scan-jobs',
				{ jobId, repoUrl, zipPath, source: repoUrl ? 'git' : file ? 'upload' : 'path' },
				{
					attempts: 3,
					backoff: { type: 'exponential', delay: 2000 },
				}
			);

			return res.status(201).json({ jobId, workspacePath });
		} catch (err) {
			console.error('[ScanController.startScan] error:', err);
			return res.status(500).json({ error: (err as Error).message });
		}
	},
	async getStatus(req: Request, res: Response) {
		try {
			const { jobId } = req.params;
			// TODO: pull job from DB and return status
			return res.json({ jobId, status: 'pending', progress: 0 });
		} catch (err) {
			return res.status(500).json({ error: (err as Error).message });
		}
	},
	async getResult(req: Request, res: Response) {
		try {
			const { jobId } = req.params;
			// TODO: fetch job result (output path or S3 URL)
			return res.json({ jobId, result: null });
		} catch (err) {
			return res.status(500).json({ error: (err as Error).message });
		}
	},
};
