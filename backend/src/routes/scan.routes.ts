import { Router } from 'express';
import multer from 'multer';
import { ScanController } from '../controllers/scan.controller';

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
router.post('/start', upload.single('file'), ScanController.startScan);
router.get('/status/:jobId', ScanController.getStatus);
router.get('/result/:jobId', ScanController.getResult);


export default router;
