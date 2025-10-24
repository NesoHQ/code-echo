import { pgTable, text, timestamp, uuid, numeric } from "drizzle-orm/pg-core";

export const jobs = pgTable("jobs", {
  id: uuid("id").defaultRandom().primaryKey(),
  userId: uuid("user_id"),
  source: text("source").notNull().default("upload"), // git|upload|path
  repoUrl: text("repo_url"),
  zipPath: text("zip_path"),
  workspacePath: text("workspace_path"),
  status: text("status").notNull().default("pending"), // pending|queued|running|done|failed
  progress: numeric("progress").notNull().default("0"),
  resultUrl: text("result_url"),
  error: text("error"),
  createdAt: timestamp("created_at").defaultNow().notNull(),
  updatedAt: timestamp("updated_at").defaultNow().notNull(),
});
