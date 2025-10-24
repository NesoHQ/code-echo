import { Queue, QueueEvents, Worker, ConnectionOptions } from 'bullmq';
import IORedis from 'ioredis';

const {
  REDIS_HOST = 'localhost',
  REDIS_PORT = '6379',
  REDIS_PASSWORD,
  BULLMQ_PREFIX = 'codeecho',
} = process.env;

// Connection configuration
export const connection: ConnectionOptions = {
  host: REDIS_HOST,
  port: Number(REDIS_PORT),
  password: REDIS_PASSWORD || undefined,
  maxRetriesPerRequest: null,
  enableReadyCheck: false,
};

// Initialize Redis client
export const redisClient = new IORedis(connection);

// Queue instance
export const scanQueue = new Queue('scan-jobs', {
  connection,
  prefix: BULLMQ_PREFIX,
});

// Queue events (logging)
const queueEvents = new QueueEvents('scan-jobs', { connection });
queueEvents.on('completed', ({ jobId }) =>
  console.log(`Job ${jobId} completed`)
);
queueEvents.on('failed', ({ jobId, failedReason }) =>
  console.error(`Job ${jobId} failed:`, failedReason)
);

// Helper to enqueue a new scan job
export async function enqueueScanJob(data: any) {
  const job = await scanQueue.add('scan', data, {
    attempts: 3,
    backoff: { type: 'exponential', delay: 3000 },
    removeOnComplete: true,
    removeOnFail: false,
  });
  console.log(`Enqueued job ${job.id}`);
  return job;
}

// Graceful shutdown helper
export async function closeQueue() {
  await queueEvents.close();
  await scanQueue.close();
  await redisClient.quit();
  console.log('BullMQ connections closed');
}
