import express from 'express';
import cors from 'cors';
import helmet from 'helmet';
import { createServer } from 'http';
import { closeQueue } from './services/queue.service';
import './workers/scan.worker';
import scanRouter from './routes/scan.routes';

const PORT = process.env.PORT ? Number(process.env.PORT) : 4000;
async function main() {
	const app = express();
	app.use(helmet());
	app.use(cors());
	app.use(express.json());

	// health
	app.get('/health', (_req, res) => {
		res.json({ status: 'ok', service: 'codeecho-backend' });
	});

	const server = createServer(app);

	server.listen(PORT, () => {
		console.log(`Backend running on port ${PORT}`);
	});

	app.use('/api/scan', scanRouter);

	const shutdown = async (signal: string) => {
		console.log(`Received ${signal}, shutting down...`);
		await closeQueue();
		server.close(() => process.exit(0));
	};

	process.on('SIGINT', () => shutdown('SIGINT'));
	process.on('SIGTERM', () => shutdown('SIGTERM'));
}

main().catch((err) => {
	console.error('Failed to start server:', err);
	process.exit(1);
});
