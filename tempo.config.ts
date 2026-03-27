import { copyFileSync, existsSync, mkdirSync, mkdtempSync, renameSync } from "node:fs";
import { tmpdir } from "node:os";
import { dirname, join } from "node:path";
import { defineConfig } from "@xevion/tempo";
import { c } from "@xevion/tempo/fmt";
import { hasTool, hasDockerDaemon, runPiped, warnMissingTool } from "@xevion/tempo/proc";
import { createOctocovConfig, testablePackages } from "@xevion/tempo/octocov";

const REPO = "Xevion/motophoto";
const CACHE_DIR = ".octocov-cache";

export default defineConfig({
  subsystems: {
    frontend: {
      aliases: ["front", "web", "fe"],
      cwd: "web",
      commands: {
        "format-check": "bunx biome check .",
        "format-apply": "bunx biome check --write .",
        lint: "bun run lint",
        "type-check": "bun run check",
        build: "bun run build",
      },
      autoFix: {
        "format-check": "format-apply",
      },
    },
    backend: {
      aliases: ["back", "go", "be"],
      requires: ["go"],
      commands: {
        "format-check": {
          cmd: 'test -z "$(goimports -l .)"',
          requires: ["goimports"],
          hint: "Run `tempo fmt backend` or `goimports -w .` to fix formatting",
        },
        "format-apply": {
          cmd: "goimports -w .",
          requires: ["goimports"],
        },
        lint: {
          cmd: "golangci-lint run --timeout=5m",
          requires: ["golangci-lint"],
        },
        build: "go build -o /dev/null .",
        test: "go test -race -count=1 ./...",
        "sqlc-diff": {
          cmd: "sqlc diff",
          requires: ["sqlc"],
          hint: "Run `just generate` to regenerate sqlc code",
        },
      },
      autoFix: {
        "format-check": "format-apply",
      },
    },
  },
  preflights: [
    (ctx) => {
      if (!existsSync("web/node_modules")) {
        ctx.fail("web/node_modules not found -- run `bun install` inside web/ first");
      }
      if (!existsSync("web/styled-system")) {
        ctx.fail("web/styled-system not found -- run `bun run codegen` inside web/ first");
      }
    },
    {
      label: "panda codegen",
      sources: { dir: "web/src", pattern: "**/*.{svelte,ts}" },
      artifacts: { dir: "web/styled-system", pattern: "**/*.{js,mjs,d.ts}" },
      regenerate: "bun run --cwd web codegen",
      reason: "svelte-check depends on styled-system types",
    },
    {
      label: "tygo bindings",
      sources: { dir: "internal/server", pattern: "types.go" },
      artifacts: { dir: "web/src/lib", pattern: "types.gen.ts" },
      regenerate: "tygo generate",
      reason: "frontend imports generated types",
    },
  ],
  check: {
    autoFixStrategy: "fix-first",
    exclude: [
      "frontend:format-apply",
      "backend:format-apply",
    ],
  },
  hooks: {
    "before:check": async (ctx) => {
      if (!hasTool("octocov")) return;

      const octocovConfig = createOctocovConfig(REPO);
      mkdirSync(".octocov", { recursive: true });
      ctx.addCleanup(octocovConfig.cleanup);
      Object.assign(ctx.env, octocovConfig.env);

      const pkgs = testablePackages();

      const sub = ctx.config.subsystems.backend;
      if (sub?.commands) {
        sub.commands.test = {
          cmd: [
            "bash", "-c",
            `go test -race -count=1 -coverprofile=coverage.out ${pkgs.join(" ")} && octocov --config=${octocovConfig.configPath} --report coverage.out`,
          ],
          warnIfExitCode: 2,
          hint: "Coverage below threshold -- run `just cov` for details.",
        };
      }
    },
    "before:dev": (ctx) => {
      const runBackend = ctx.targets.has("backend");
      const runFrontend = ctx.targets.has("frontend");

      if (runFrontend && !existsSync("web/node_modules")) {
        ctx.fail("web/node_modules not found -- run `bun install` inside web/ first");
      }

      if (runBackend) {
        if (!existsSync(".env")) {
          ctx.fail(".env not found -- copy .env.example first: `cp .env.example .env`");
        }
        if (!hasTool("docker")) {
          ctx.logger.warn("docker not found -- if you use Docker for PostgreSQL, install it first");
        } else if (!hasDockerDaemon()) {
          ctx.logger.warn("Docker daemon is not running -- run `just db` to start PostgreSQL");
        } else {
          const dbStatus = runPiped(
            ["docker", "compose", "ps", "--status", "running", "--quiet", "db"],
          );
          const dbRunning =
            dbStatus.exitCode === 0 &&
            dbStatus.stdout.trim().length > 0;
          if (!dbRunning) {
            ctx.logger.warn("Database container is not running -- run `just db` to start PostgreSQL");
          }
        }
        if (!hasTool("air")) {
          ctx.fail("air not found -- cannot start backend. Install air first.");
        }
      }
    },
  },
  dev: {
    exitBehavior: "first-exits",
    processes: {
      frontend: {
        type: "unmanaged",
        cmd: ["bun", "run", "dev"],
        cwd: "web",
      },
      backend: {
        type: "unmanaged",
        cmd: ["air", "-build.send_interrupt", "true"],
        env: { PORT: "3001" },
      },
    },
  },
  custom: {
    cov: async ({ run }) => {
      if (!hasTool("octocov")) {
        warnMissingTool("octocov", "cannot run coverage report");
        return 1;
      }

      // Fetch the latest successful master CI coverage artifact as a diff baseline,
      // cached by commit SHA so subsequent runs are instant.
      if (hasTool("gh")) {
        const authOk =
          runPiped(["gh", "auth", "status"]).exitCode === 0;

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
            const localReport = `.octocov/${REPO}/report.json`;

            if (!existsSync(cached)) {
              process.stdout.write(`${c.dim(`Fetching coverage baseline (master@${runSha.slice(0, 7)})...`)}\n`);
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
                process.stdout.write(`${c.yellow("Baseline unavailable")} -- running without diff\n`);
              }
            }

            if (existsSync(cached)) {
              mkdirSync(dirname(localReport), { recursive: true });
              copyFileSync(cached, localReport);
            }
          }
        }
      }

      const { configPath, env, cleanup } = createOctocovConfig(REPO);
      mkdirSync(".octocov", { recursive: true });

      try {
        run(["go", "test", "-race", "-count=1", "-coverprofile=coverage.out", "./..."]);
        run(["octocov", "--config", configPath, "--report", "coverage.out"], { env });
      } finally {
        cleanup();
      }
      return 0;
    },
  },
});
