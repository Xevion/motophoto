# Architecture

MotoPhoto is an event photography marketplace — photographers upload action sports photos, customers browse and purchase them. The system is a monorepo with a Go backend serving a SvelteKit frontend.

## System Overview

```
┌─────────────────────────────────────────────────┐
│                   Client Browser                 │
│                                                  │
│  SvelteKit (SSR) ──► Go API ──► PostgreSQL       │
│       :5173              :3001       :57512       │
└─────────────────────────────────────────────────┘
```

| Component | Tech | Port | Directory |
|-----------|------|------|-----------|
| Frontend | SvelteKit 2 + Svelte 5 | 5173 (dev) | `web/` |
| Backend | Go + Chi router | 3001 | `main.go`, `internal/` |
| Database | PostgreSQL 17 | 57512 (local) | `docker-compose.yml` |

## Request Flow

### Development

Two servers run simultaneously via `just dev`:

1. **SvelteKit dev server** (Vite, port 5173) — serves the frontend with HMR
2. **Go backend** (Air hot-reload, port 3001) — serves the API

Vite proxies `/api` requests to the Go backend (`web/vite.config.ts`). The browser only talks to `:5173`.

```
Browser ──► Vite (:5173)
              ├── static assets, HMR ──► browser
              └── /api/* proxy ──► Go (:3001) ──► Postgres
```

### Production

The container runs two processes orchestrated by `web/entrypoint.ts`:

1. **Go backend** (port 3001, internal only) — serves the API
2. **SvelteKit SSR** (port $PORT, public) — server-side renders pages

SvelteKit's `hooks.server.ts` forwards `/api/*` requests to the Go backend. The entrypoint starts Go first, waits for a health check (`/api/health`), then starts SvelteKit.

```
Browser ──► SvelteKit SSR (:$PORT)
              ├── SSR page rendering
              └── /api/* proxy (hooks.server.ts) ──► Go (:3001) ──► Postgres
```

## Project Structure

```
motophoto/
├── main.go                          # Entry point — loads env, creates logger, starts server
├── internal/
│   ├── server/
│   │   └── server.go                # Chi router setup, middleware, route definitions
│   ├── database/
│   │   ├── database.go              # pgx connection pool creation
│   │   ├── migrations/              # SQL migration files (sequential, numbered)
│   │   ├── queries/                 # sqlc query definitions (.sql)
│   │   └── db/                      # sqlc generated Go code (DO NOT EDIT)
│   └── middleware/
│       └── middleware.go            # Custom middleware (placeholder)
├── web/                             # SvelteKit frontend (see SVELTE.md)
│   ├── src/
│   │   ├── routes/                  # File-based routing
│   │   └── lib/                     # Shared code (api.ts, components)
│   ├── svelte.config.js
│   └── vite.config.ts               # API proxy config
├── Justfile                         # Task runner — all dev commands
├── scripts/                         # Bun-based dev scripts (check, dev)
│   ├── check.ts                     # Parallel check runner
│   ├── dev.ts                       # Dev server orchestrator
│   └── lib/                         # Shared utilities (fmt, proc)
├── docker-compose.yml               # Local Postgres
├── Dockerfile                       # Multi-stage production build
├── sqlc.yml                         # SQL code generation config
├── .air.toml                        # Go hot-reload config
└── .github/workflows/ci.yml        # Lint, test, Docker build
```

## Tools

| Tool | Purpose | Required |
|------|---------|----------|
| [Go 1.26+](https://go.dev) | Backend language | Yes |
| [Bun](https://bun.sh) | Frontend runtime, script runner | Yes |
| [just](https://just.systems) | Task runner | Yes |
| [Air](https://github.com/air-verse/air) | Go hot-reload dev server | For `just dev` |
| [golangci-lint](https://golangci-lint.run) | Go linter | No — skipped if missing |
| [tygo](https://github.com/gzuidhof/tygo) | Go → TypeScript type generation | No — skipped if missing |
| [sqlc](https://sqlc.dev) | SQL → Go code generation | For `just generate` |
| [Docker](https://docs.docker.com/engine/install/) | Local Postgres | For `just db` |

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | (see `.env.example`) | PostgreSQL connection string |
| `PORT` | `3001` | Go server listen port |

## Database

PostgreSQL 17 runs locally via docker-compose on port **57512** (mapped from container's 5432).

Connection string: `postgres://motophoto:motophoto@localhost:57512/motophoto`

The database layer (pgx pool + sqlc generated queries) exists but is **not yet wired into the API handlers** — current endpoints return hardcoded demo data. See [GO.md](GO.md) for the sqlc workflow.

## Deployment

### Docker

The `Dockerfile` uses a 3-stage build:

1. **Go stage** (`golang:1.26-alpine`) — compiles the Go binary, compresses with UPX
2. **Frontend stage** (`oven/bun:1`) — builds SvelteKit with `bun run build`
3. **Runtime stage** (`oven/bun:1-slim`) — Bun runtime with the Go binary, SvelteKit build output, and `web/entrypoint.ts` as the entrypoint

The runtime container runs `bun run /app/web/entrypoint.ts`, which starts the Go backend on port 3001, waits for health, then starts SvelteKit SSR on the public port.

### CI/CD

GitHub Actions (`.github/workflows/ci.yml`) runs on every push:

1. **lint-and-test** — Postgres service container, golangci-lint, Go tests, svelte-check, ESLint
2. **docker** (master only) — builds and pushes to GitHub Container Registry (GHCR)

### Railway

Production deploys to Railway, which reads the `Dockerfile` directly. The `PORT` env var is set by Railway.
