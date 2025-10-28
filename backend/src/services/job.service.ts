import { db } from '../db';
import { jobs } from '../db/schema/jobs';
import { eq } from 'drizzle-orm';

export type JobStatus = 'pending' | 'queued' | 'running' | 'done' | 'failed';

interface CreateJobInput {
	id: string;
	userId?: string;
	source?: 'git' | 'upload' | 'path';
	repoUrl?: string;
	zipPath?: string;
	workspacePath?: string;
}

/**
 * Create a new job record in the database.
 */
export async function createJob(input: CreateJobInput) {
	const now = new Date();

	await db.insert(jobs).values({
		id: input.id,
		userId: input.userId ?? null,
		source: input.source ?? 'upload',
		repoUrl: input.repoUrl,
		zipPath: input.zipPath,
		workspacePath: input.workspacePath,
		status: 'queued',
		progress: '0',
		createdAt: now,
		updatedAt: now,
	});
}

/**
 * Update a job's status and optional result/error fields.
 */
export async function updateJobStatus(
	id: string,
	status: JobStatus,
	data?: { resultUrl?: string; error?: string; progress?: string }
) {
	const now = new Date();

	await db
		.update(jobs)
		.set({
			status,
			resultUrl: data?.resultUrl,
			error: data?.error,
			progress: data?.progress ?? '0',
			updatedAt: now,
		})
		.where(eq(jobs.id, id));
}

/**
 * Fetch a single job by ID.
 */
export async function getJobById(id: string) {
	const [job] = await db.select().from(jobs).where(eq(jobs.id, id));
	return job;
}

/**
 * List recent jobs.
 */
export async function listJobs(limit = 10) {
	return await db.select().from(jobs).limit(limit);
}
