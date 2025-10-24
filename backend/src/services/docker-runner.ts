// src/services/docker-runner.ts
import { spawn } from "child_process";
import path from "path";
import fs from "fs";

export async function runCodeEchoCLI({ workspace, jobId }: { workspace: string; jobId: string }) {
  const outputFile = path.join(workspace, "output.xml");

  console.log(`🚀 Running CodeEcho CLI natively for job ${jobId}...`);

  return new Promise<void>((resolve, reject) => {
    const proc = spawn("codeecho", ["scan", workspace, "--format", "xml", "-o", outputFile], {
      stdio: ["ignore", "pipe", "pipe"],
    });

    proc.stdout.on("data", (data) => console.log(`[job ${jobId}] ${data.toString().trim()}`));
    proc.stderr.on("data", (data) => console.error(`[job ${jobId} ERROR] ${data.toString().trim()}`));

    proc.on("close", (code) => {
      if (code === 0 && fs.existsSync(outputFile)) {
        console.log(`✅ CodeEcho scan finished for job ${jobId}`);
        resolve();
      } else {
        reject(new Error(`CodeEcho CLI failed with exit code ${code}`));
      }
    });

    proc.on("error", (err) => reject(err));
  });
}
