/**
 * Helpers for running octocov locally with a patched config.
 *
 * octocov's artifact:// datastore only works inside GitHub Actions. For local
 * runs we swap every artifact:// entry with local://.octocov so the report is
 * read from / written to disk instead.
 */

import { mkdtempSync, readFileSync, rmSync, unlinkSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

export const REPO = "Xevion/motophoto";
export const LOCAL_STORE = "local://.octocov";
export const ARTIFACT_STORE = "artifact://${GITHUB_REPOSITORY}";

export interface LocalConfig {
	/** Path to the patched .octocov.yml temp file. */
	configPath: string;
	/** Path to a minimal GitHub event JSON file required by octocov. */
	eventPath: string;
	/** Environment variables that must be set when invoking octocov. */
	env: Record<string, string>;
	/** Remove the temp directory when done. */
	cleanup: () => void;
}

/**
 * Read sourcePath (.octocov.yml), replace every artifact:// datastore entry
 * with local://.octocov, and write the result to a temp directory alongside a
 * minimal GitHub event JSON. Returns paths and the env block needed by octocov.
 */
export function createLocalConfig(sourcePath = ".octocov.yml"): LocalConfig {
	const text = readFileSync(sourcePath, "utf-8");
	const patched = text.replaceAll(ARTIFACT_STORE, LOCAL_STORE);

	// Config must live in the project root so octocov resolves coverage.out,
	// codeToTestRatio globs, and local:// datastore paths relative to CWD.
	const configPath = ".octocov-local.yml";

	// Event file can live in /tmp — its path is passed via env, not config.
	const eventDir = mkdtempSync(join(tmpdir(), "octocov-event-"));
	const eventPath = join(eventDir, "event.json");

	writeFileSync(configPath, patched, "utf-8");
	writeFileSync(eventPath, "{}", "utf-8");

	return {
		configPath,
		eventPath,
		env: {
			GITHUB_REPOSITORY: REPO,
			GITHUB_EVENT_NAME: "push",
			GITHUB_EVENT_PATH: eventPath,
		},
		cleanup: () => {
			try { unlinkSync(configPath); } catch {}
			try { rmSync(eventDir, { recursive: true }); } catch {}
		},
	};
}
