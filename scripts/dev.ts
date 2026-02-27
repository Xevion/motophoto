/**
 * Dev server orchestrator for MotoPhoto.
 *
 * Usage: bun scripts/dev.ts [flags]
 *
 * Flags:
 *   -f, --frontend-only   Frontend only (Vite dev server)
 *   -b, --backend-only    Backend only (Air hot-reload)
 */

import { existsSync } from "node:fs";
import { parseFlags, c } from "./lib/fmt";
import { ProcessGroup, assertNotWindowsNative, hasDockerDaemon, hasTool, warnMissingTool } from "./lib/proc";

assertNotWindowsNative();

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

// Pre-flight: frontend requires web/node_modules
if (runFrontend && !existsSync("web/node_modules")) {
	process.stderr.write(
		`${c("31", "✗ web/node_modules not found")} — run \`bun install\` inside web/ first\n`,
	);
	process.exit(1);
}

// Pre-flight: backend requires .env and a reachable database
if (runBackend) {
	if (!existsSync(".env")) {
		process.stderr.write(
			`${c("31", "✗ .env not found")} — copy .env.example first: \`cp .env.example .env\`\n`,
		);
		process.exit(1);
	}

	if (!hasTool("docker")) {
		process.stderr.write(
			`${c("33", "⚠ docker not found")} — if you use Docker for PostgreSQL, install it first — see README.md\n`,
		);
	} else if (!hasDockerDaemon()) {
		process.stderr.write(
			`${c("33", "⚠ Docker daemon is not running")} — start Docker, then run \`just db\` to start PostgreSQL\n`,
		);
	} else {
		const dbStatus = Bun.spawnSync(
			["docker", "compose", "ps", "--status", "running", "--quiet", "db"],
			{ stdout: "pipe", stderr: "pipe" },
		);
		const dbRunning = dbStatus.exitCode === 0 && dbStatus.stdout.toString().trim().length > 0;
		if (!dbRunning) {
			process.stderr.write(
				`${c("33", "⚠ Database container is not running")} — run \`just db\` to start PostgreSQL\n`,
			);
		}
	}
}

const group = new ProcessGroup();

if (runFrontend) {
	console.log(c("1;36", "→ Starting frontend dev server..."));
	group.spawn(["bun", "run", "--cwd", "web", "dev"]);
}

if (runBackend) {
	if (!hasTool("air")) {
		warnMissingTool("air", backendOnly ? "cannot start backend" : "skipping backend");
		if (backendOnly) process.exit(1);
	} else {
		console.log(c("1;36", "→ Starting backend dev server (Air)..."));
		group.spawn(["air", "-build.send_interrupt", "true"], {
			env: { ...process.env, PORT: "3001" },
		});
	}
}

const code = await group.waitForFirst();
process.exit(code);
