import { db } from "../db";
import { jobs } from "../db/schema/jobs";
import { eq } from "drizzle-orm";

export async function createJob(payload: Partial<typeof jobs.$inferInsert>) {
  const [job] = await db.insert(jobs).values(payload).returning();
  return job;
}

export async function updateJobStatus(
  id: string,
  updates: Partial<typeof jobs.$inferInsert>
) {
  const [job] = await db.update(jobs).set(updates).where(eq(jobs.id, id)).returning();
  return job;
}

export async function getJobById(id: string) {
  const [job] = await db.select().from(jobs).where(eq(jobs.id, id));
  return job;
}

export async function listJobs(limit = 10) {
  return await db.select().from(jobs).limit(limit);
}
