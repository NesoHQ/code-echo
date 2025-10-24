import { Worker, Job } from "bullmq";
import { connection, closeQueue } from "../services/queue.service";
import { runCodeEchoCLI } from "../services/docker-runner";
import path from "path";
import fs from "fs";
import os from "os";
import { v4 as uuidv4 } from "uuid";

interface ScanJobData {
  repoUrl?: string;
  zipPath?: string;
}

const QUEUE_NAME = "scan-jobs";
const CONCURRENCY = Number(process.env.WORKER_CONCURRENCY ?? 2);
const PREFIX = process.env.BULLMQ_PREFIX ?? "codeecho";

/**
 * Process a single scan job.
 * Defensive: job.id may be undefined according to types, so create a stable string id to use.
 */
async function processScanJob(job: Job<ScanJobData>) {
  const jobId = job.id?.toString() ?? uuidv4();
  console.log(`Worker started job ${jobId} (internal job.id=${String(job.id)})`);
  console.log("Job data:", job.data);

  // Prepare workspace
  const workspace = path.resolve(os.tmpdir(), `codeecho-job-${jobId}`);
  try {
    await fs.promises.mkdir(workspace, { recursive: true });
  } catch (err) {
    console.error(`[${jobId}] failed to create workspace ${workspace}:`, err);
    throw err;
  }

  try {
    // Run the CLI (this will stream logs in docker-runner)
    await runCodeEchoCLI({ workspace, jobId });

    const outputPath = path.join(workspace, "output.xml");
    if (!(await exists(outputPath))) {
      throw new Error(`Expected output missing at ${outputPath}`);
    }

    console.log(`Job ${jobId} done - result stored at ${outputPath}`);
    return { status: "done", output: outputPath };
  } catch (rawErr) {
    const errMsg = rawErr instanceof Error ? rawErr.message : String(rawErr);
    console.error(`Job ${jobId} failed:`, errMsg);
    // Keep the error so BullMQ can register failure and retries
    throw new Error(errMsg);
  } finally {
    // Optionally keep workspace for debugging. If you want to cleanup, uncomment:
    // try { await fs.promises.rm(workspace, { recursive: true, force: true }); } catch (e) { /* log if needed */ }
  }
}

// helper
async function exists(p: string) {
  try {
    await fs.promises.access(p);
    return true;
  } catch {
    return false;
  }
}

// Create worker instance
export const scanWorker = new Worker<ScanJobData>(QUEUE_NAME, processScanJob, {
  connection,
  concurrency: CONCURRENCY,
  prefix: PREFIX,
});

// Correct event signatures per bullmq:
// completed(jobId, returnValue), failed(jobId, failedReason), error(err)
scanWorker.on("completed", (jobId, returnValue) => {
  console.log(`Job ${String(jobId)} completed. result:`, returnValue);
});

scanWorker.on("failed", (jobId, failedReason) => {
  console.error(`Job ${String(jobId)} failed:`, failedReason);
});

scanWorker.on("error", (err) => {
  console.error("Worker error:", err instanceof Error ? err.message : String(err));
});

console.log(`Scan worker initialized (queue=${QUEUE_NAME}, prefix=${PREFIX}, concurrency=${CONCURRENCY})`);

// Graceful shutdown with timeout
async function shutdown(signal: string) {
  console.log(`Received ${signal}, closing worker...`);
  try {
    // close worker (stop processing new jobs and wait for active jobs to finish)
    await scanWorker.close();
    // close queue related resources
    await closeQueue();
    console.log("Worker and queue closed gracefully.");
    // allow logs to flush
    setTimeout(() => process.exit(0), 250);
  } catch (err) {
    console.error("Error during shutdown:", err);
    // force exit after short timeout
    setTimeout(() => process.exit(1), 250);
  }
}

process.on("SIGINT", () => shutdown("SIGINT"));
process.on("SIGTERM", () => shutdown("SIGTERM"));