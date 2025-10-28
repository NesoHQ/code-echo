import { Router, Request, Response } from 'express';
import multer from 'multer';
import path from 'path';
import fs from 'fs';
import { v4 as uuidv4 } from 'uuid';
import { createJob } from '../services/job.service';
import { sendToQueue } from '../services/queue.service';

const upload = multer({
	dest: '/tmp/uploads',
	limits: { fileSize: Number(process.env.MAX_ZIP_SIZE_BYTES ?? 209715200) },
});
const router = Router();

/**
 * POST /api/scan/start
 * Body JSON: { repoUrl?: string, path?: string }
 * OR multipart/form-data with field "file" = zip
 */
router.post('/start', upload.single('file'), async (req: Request, res: Response) => {
	try {
		const repoUrl = (req.body.repoUrl || req.query.repoUrl) as string | undefined;
		const localPath = (req.body.path || req.query.path) as string | undefined;
		const file = req.file; // optional zip

		if (!repoUrl && !file && !localPath) {
			return res.status(400).json({ error: 'Provide repoUrl or upload a zip file or a local path' });
		}

		const jobId = uuidv4();
		const workspacePath = path.join(process.env.TEMP_DIR ?? '/app/data/outputs', jobId);

		// create workspace directory
		await fs.promises.mkdir(workspacePath, { recursive: true });

		// if uploaded file, move it into workspace as upload.zip
		let zipPath: string | undefined;
		if (file) {
			const dest = path.join(workspacePath, 'upload.zip');
			await fs.promises.rename(file.path, dest);
			zipPath = dest;
		}

		// create DB record
		await createJob({
			id: jobId,
			source: repoUrl ? 'git' : file ? 'upload' : 'path',
			repoUrl,
			zipPath,
			workspacePath,
		});

		// enqueue job into BullMQ: include jobId and the source info
		await sendToQueue(
			'scan-jobs',
			{ jobId, repoUrl, zipPath, source: repoUrl ? 'git' : file ? 'upload' : 'path' },
			{
				attempts: 3,
				backoff: { type: 'exponential', delay: 2000 },
			}
		);

		res.status(201).json({ jobId, workspacePath });
	} catch (err) {
		console.error('scan.start error:', err);
		res.status(500).json({ error: (err as Error).message });
	}
});

export default router;
