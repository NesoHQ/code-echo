CREATE TABLE "jobs" (
	"id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"user_id" uuid,
	"source" text DEFAULT 'upload' NOT NULL,
	"repo_url" text,
	"zip_path" text,
	"workspace_path" text,
	"status" text DEFAULT 'pending' NOT NULL,
	"progress" numeric DEFAULT '0' NOT NULL,
	"result_url" text,
	"error" text,
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp DEFAULT now() NOT NULL
);
