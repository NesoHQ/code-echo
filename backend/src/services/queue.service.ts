import { Queue, QueueOptions } from 'bullmq';
import IORedis, { RedisOptions } from 'ioredis';

const { REDIS_HOST = 'localhost', REDIS_PORT = '6379', REDIS_PASSWORD, BULLMQ_PREFIX = 'codeecho' } = process.env;

// Connection configuration
export const redisConfig: RedisOptions = {
	host: REDIS_HOST,
	port: Number(REDIS_PORT),
	password: REDIS_PASSWORD || undefined,
	maxRetriesPerRequest: null,
	enableReadyCheck: false,
	retryStrategy: (times) => Math.min(times * 50, 2000),
};

export const connection = new IORedis(redisConfig);

const queues: Record<string, Queue> = {};

export function getQueue(name: string) {
	if (!queues[name]) {
		queues[name] = new Queue(name, { connection, prefix: process.env.BULLMQ_PREFIX ?? 'codeecho' });
	}
	return queues[name];
}

export async function sendToQueue(name: string, payload: Record<string, any>, opts?: QueueOptions & any) {
	const q = getQueue(name);
	const job = await q.add(name, payload, opts);
	console.log(`Enqueued job ${job.id} -> queue=${name}`);
	return job;
}

export async function enqueueScanJob(data: Record<string, any>) {
	return sendToQueue('scan-jobs', data, {
		attempts: 3,
		backoff: { type: 'exponential', delay: 3000 },
		removeOnComplete: true,
		removeOnFail: false,
	});
}

export async function closeQueue() {
	try {
		await Promise.all(Object.values(queues).map((q) => q.close()));
	} catch (e) {
		console.error('Error closing queues:', e);
	}
	try {
		await connection.quit();
	} catch (e) {
		console.error('Error quitting Redis:', e);
	}
	console.log('BullMQ connections closed');
}
