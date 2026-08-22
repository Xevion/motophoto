import {
  copyFileSync,
  existsSync,
  mkdirSync,
  mkdtempSync,
  readFileSync,
  renameSync,
  rmSync,
  unlinkSync,
  writeFileSync,
} from "node:fs";
import { tmpdir } from "node:os";
import { dirname, join } from "node:path";
import { defineConfig, presets, task } from "@xevion/tempo";
import type { RunContext } from "@xevion/tempo";

const REPO = "Xevion/motophoto";
const CACHE_DIR = ".octocov-cache";
// biome-ignore lint/suspicious/noTemplateCurlyInString: the literal string octocov expects
const ARTIFACT_STORE = "artifact://${GITHUB_REPOSITORY}";

/** Point octocov at a local datastore instead of the CI artifact one. */
function localOctocov(root: string) {
  const configPath = join(root, ".octocov-local.yml");
  const eventDir = mkdtempSync(join(tmpdir(), "octocov-event-"));
  const eventPath = join(eventDir, "event.json");

  const source = readFileSync(join(root, ".octocov.yml"), "utf-8");
  writeFileSync(configPath, source.replaceAll(ARTIFACT_STORE, "local://.octocov"), "utf-8");
  writeFileSync(eventPath, "{}", "utf-8");
  mkdirSync(join(root, ".octocov"), { recursive: true });

  return {
    configPath,
    env: {
      GITHUB_REPOSITORY: REPO,
      GITHUB_EVENT_NAME: "push",
      GITHUB_EVENT_PATH: eventPath,
    },
    cleanup: () => {
      try {
        unlinkSync(configPath);
      } catch {}
      rmSync(eventDir, { recursive: true, force: true });
    },
  };
}

/** The most recent successful master coverage report, cached by commit. */
async function fetchBaseline(ctx: RunContext): Promise<void> {
  if ((await ctx.capture(["gh", "auth", "status"])).code !== 0) return;

  const listed = await ctx.capture([
    "gh", "run", "list",
    "--branch", "master",
    "--workflow", "ci.yml",
    "--status", "success",
    "--limit", "1",
    "--json", "databaseId,headSha",
  ]);
  const runs: { databaseId: number; headSha: string }[] = JSON.parse(listed.stdout || "[]");
  const head = runs[0];
  if (!head) return;

  const cached = join(ctx.cwd, CACHE_DIR, `${head.headSha}.json`);
  if (!existsSync(cached)) {
    ctx.log(`fetching coverage baseline (master@${head.headSha.slice(0, 7)})`);
    mkdirSync(join(ctx.cwd, CACHE_DIR), { recursive: true });
    const tmp = mkdtempSync(join(tmpdir(), "octocov-baseline-"));
    const code = await ctx.run([
      "gh", "run", "download", String(head.databaseId),
      "--name", "octocov-report", "--dir", tmp,
    ]);
    if (code !== 0) {
      ctx.log("baseline unavailable -- running without a diff");
      return;
    }
    renameSync(join(tmp, "report.json"), cached);
  }

  const report = join(ctx.cwd, ".octocov", REPO, "report.json");
  mkdirSync(dirname(report), { recursive: true });
  copyFileSync(cached, report);
}

export default defineConfig({
  tasks: [
    task({
      name: "web:deps",
      cwd: "web",
      body: (ctx) => {
        if (!existsSync(join(ctx.cwd, "node_modules"))) {
          ctx.fail("web/node_modules not found -- run `bun install` inside web/ first");
        }
      },
    }),

    // svelte-check reads the generated styled-system types.
    task({
      name: "web:codegen",
      cwd: "web",
      body: "bun run codegen",
      needs: ["web:deps"],
      inputs: ["web/panda.config.ts", "web/src/**/*.{svelte,ts}"],
      outputs: ["web/styled-system/**/*.{js,mjs,d.ts}"],
    }),

    // The frontend imports these generated types.
    task({
      name: "web:bindings",
      body: "tygo generate",
      requires: [{ tool: "tygo" }],
      inputs: ["tygo.yaml", "internal/server/types.go"],
      outputs: ["web/src/lib/types.gen.ts"],
    }),

    task({
      name: "web:format",
      cwd: "web",
      body: "bunx biome check .",
      tags: ["check"],
      needs: ["web:deps"],
    }),
    task({
      name: "web:format-fix",
      cwd: "web",
      body: "bunx biome check --write .",
      tags: ["format"],
      needs: ["web:deps"],
    }),
    task({
      name: "web:lint",
      cwd: "web",
      body: "bun run lint",
      tags: ["check", "lint"],
      needs: ["web:deps"],
    }),
    task({
      name: "web:type-check",
      cwd: "web",
      body: "bun run check",
      tags: ["check"],
      needs: ["web:deps", "web:codegen", "web:bindings"],
    }),
    task({
      name: "web:build",
      cwd: "web",
      body: "bun run build",
      tags: ["check", "build"],
      needs: ["web:deps", "web:codegen", "web:bindings"],
    }),

    ...presets.go({
      name: "backend",
      override: { build: "go build -o /dev/null ." },
    }),
    task({
      name: "backend:sqlc-diff",
      body: "sqlc diff",
      tags: ["check"],
      requires: [{ tool: "sqlc" }],
    }),

    task({
      name: "backend:coverage",
      tags: ["cov"],
      requires: [{ tool: "octocov" }, { tool: "gh" }],
      body: async (ctx) => {
        await fetchBaseline(ctx);
        const { configPath, env, cleanup } = localOctocov(ctx.cwd);
        try {
          // Only packages with tests, so 0%-coverage noise stays out of the report.
          const listed = await ctx.capture([
            "go", "list", "-f",
            "{{if or .TestGoFiles .XTestGoFiles}}{{.ImportPath}}{{end}}",
            "./...",
          ]);
          const pkgs = listed.stdout.trim().split("\n").filter(Boolean);
          const test = await ctx.run([
            "go", "test", "-race", "-count=1", "-coverprofile=coverage.out", ...pkgs,
          ]);
          if (test !== 0) return test;
          for (const [key, value] of Object.entries(env)) process.env[key] = value;
          return await ctx.run(["octocov", "--config", configPath, "--report", "coverage.out"]);
        } finally {
          cleanup();
        }
      },
    }),

    // Warns rather than blocks: a missing database is a dev inconvenience, not a stop.
    task({
      name: "backend:db-ready",
      tags: ["dev"],
      requires: [{ tool: "docker" }],
      body: async (ctx) => {
        const ps = await ctx.capture(["docker", "compose", "ps", "--status", "running", "--quiet", "db"]);
        if (ps.code !== 0 || ps.stdout.trim() === "") {
          ctx.log("database container is not running -- run `just db`");
        }
      },
    }),

    task({
      name: "web:dev",
      cwd: "web",
      body: ["bun", "run", "dev"],
      tags: ["dev"],
      persistent: true,
      needs: ["web:deps"],
      passthrough: true,
    }),
    task({
      name: "backend:dev",
      body: ["air", "-build.send_interrupt", "true"],
      tags: ["dev"],
      persistent: true,
      env: { PORT: "3001" },
      requires: [
        { tool: "air" },
        { file: ".env", hint: "cp .env.example .env" },
      ],
      // Ordering only: a skipped database check must not block the server.
      after: ["backend:db-ready"],
    }),
  ],

  commands: {
    check: { description: "Run every check", tags: ["check"] },
    fmt: { description: "Apply every formatter", tags: ["format"], concurrency: 1 },
    lint: { description: "Lint every subsystem", tags: ["lint"] },
    dev: { description: "Run the dev servers", tags: ["dev"], exitBehavior: "first-exits" },
    cov: { description: "Go coverage, diffed against the master baseline", tags: ["cov"] },
  },
});
