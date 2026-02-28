/**
 * Run go test with coverage and octocov diff against the master baseline.
 *
 * 1. Fetch the latest successful master CI artifact as a diff baseline (cached
 *    by commit SHA in .octocov-cache/ so subsequent runs are instant).
 * 2. Patch .octocov.yml to use local://.octocov instead of artifact://.
 * 3. Run go test + octocov with the patched config.
 *
 * Usage: bun scripts/octocov-local.ts
 */

import { copyFileSync, existsSync, mkdirSync, mkdtempSync, renameSync } from "node:fs";
import { tmpdir } from "node:os";
import { dirname, join } from "node:path";
import { c } from "./lib/fmt";
import { createLocalConfig, REPO } from "./lib/octocov";
import { hasTool, run, runPiped, warnMissingTool } from "./lib/proc";

const CACHE_DIR = ".octocov-cache";
const LOCAL_REPORT = `.octocov/${REPO}/report.json`;

if (!hasTool("octocov")) {
	warnMissingTool("octocov", "cannot run coverage report");
	process.exit(1);
}

// Attempt to fetch the master baseline from the last successful CI run.
// Requires `gh` with a valid auth token; silently skips when unavailable.
if (hasTool("gh")) {
	const authOk =
		Bun.spawnSync(["gh", "auth", "status"], { stdout: "pipe", stderr: "pipe" }).exitCode === 0;

	if (authOk) {
		const listResult = runPiped([
			"gh", "run", "list",
			"--branch", "master",
			"--workflow", "ci.yml",
			"--status", "success",
			"--limit", "1",
			"--json", "databaseId,headSha",
		]);

		const runs: { databaseId: number; headSha: string }[] = JSON.parse(
			listResult.stdout || "[]",
		);
		const runId = runs[0]?.databaseId;
		const runSha = runs[0]?.headSha;

		if (runId && runSha) {
			const cached = join(CACHE_DIR, `${runSha}.json`);

			if (!existsSync(cached)) {
				process.stdout.write(
					`${c("2", `Fetching coverage baseline (master@${runSha.slice(0, 7)})...`)}\n`,
				);
				mkdirSync(CACHE_DIR, { recursive: true });

				const tmp = mkdtempSync(join(tmpdir(), "octocov-baseline-"));
				const dlResult = runPiped([
					"gh", "run", "download", String(runId),
					"--name", "octocov-report",
					"--dir", tmp,
				]);

				if (dlResult.exitCode === 0) {
					renameSync(join(tmp, "report.json"), cached);
				} else {
					process.stdout.write(`${c("33", "Baseline unavailable")} -- running without diff\n`);
				}
			}

			if (existsSync(cached)) {
				mkdirSync(dirname(LOCAL_REPORT), { recursive: true });
				copyFileSync(cached, LOCAL_REPORT);
			}
		}
	}
}

// Patch .octocov.yml to use local:// and set up a temp event file.
const { configPath, env, cleanup } = createLocalConfig();
process.on("exit", cleanup);
process.on("SIGINT", () => {
	cleanup();
	process.exit(130);
});

// Ensure the local datastore dir exists so octocov doesn't fail on stat.
mkdirSync(".octocov", { recursive: true });

run(["go", "test", "-race", "-count=1", "-coverprofile=coverage.out", "./..."]);
run(["octocov", "--config", configPath, "--report", "coverage.out"], { env });
