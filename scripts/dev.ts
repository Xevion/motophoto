/**
 * Dev server orchestrator for MotoPhoto.
 *
 * Usage: bun scripts/dev.ts [flags]
 *
 * Flags:
 *   -f, --frontend-only   Frontend only (Vite dev server)
 *   -b, --backend-only    Backend only (Air hot-reload)
 */

import { parseFlags, c } from "./lib/fmt";
import { ProcessGroup } from "./lib/proc";

const { flags } = parseFlags(
	process.argv.slice(2),
	{
		"frontend-only": "bool",
		"backend-only": "bool",
	} as const,
	{
		f: "frontend-only",
		b: "backend-only",
	},
	{
		"frontend-only": false,
		"backend-only": false,
	},
);

const frontendOnly = flags["frontend-only"];
const backendOnly = flags["backend-only"];

const anySpecified = frontendOnly || backendOnly;
const runFrontend = anySpecified ? frontendOnly : true;
const runBackend = anySpecified ? backendOnly : true;

const group = new ProcessGroup();

if (runFrontend) {
	console.log(c("1;36", "→ Starting frontend dev server..."));
	group.spawn(["bun", "run", "--cwd", "web", "dev"]);
}

if (runBackend) {
	console.log(c("1;36", "→ Starting backend dev server (Air)..."));
	group.spawn(["air", "-build.send_interrupt", "true"]);
}

const code = await group.waitForFirst();
process.exit(code);
