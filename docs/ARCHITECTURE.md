# Architecture

MotoPhoto is an event photography marketplace — photographers upload action sports photos, customers browse and purchase them. The system is a monorepo with a Go backend serving a SvelteKit frontend.

## System Overview

```
┌─────────────────────────────────────────────────┐
│                   Client Browser                 │
│                                                  │
│  SvelteKit (SSR) ──► Go API ──► PostgreSQL       │
│       :5173              :8080       :57512       │
└─────────────────────────────────────────────────┘
```

| Component | Tech | Port | Directory |
|-----------|------|------|-----------|
| Frontend | SvelteKit 2 + Svelte 5 | 5173 (dev) | `web/` |
| Backend | Go + Chi router | 8080 | `main.go`, `internal/` |
| Database | PostgreSQL 17 | 57512 (local) | `docker-compose.yml` |

## Request Flow

### Development

In development, two servers run simultaneously via `task dev`:

1. **SvelteKit dev server** (Vite, port 5173) — serves the frontend with HMR
2. **Go backend** (Air hot-reload, port 8080) — serves the API

Vite proxies `/api` and `/health` requests to the Go backend (`web/vite.config.ts`). This means the browser only talks to `:5173`, and Vite forwards API calls transparently.

For SSR (server-side rendering), SvelteKit's `+page.server.ts` load functions call the Go API directly at `http://localhost:8080` using the `apiFetch` helper in `web/src/lib/api.ts`.

```
Browser ──► Vite (:5173)
              ├── static assets, HMR ──► browser
              └── /api/* proxy ──► Go (:8080) ──► Postgres
```

### Production

In production, the Go binary serves both the API and the pre-built SvelteKit static output. There's no separate frontend server — the Go backend handles everything behind a single port.

```
Browser ──► Go (:$PORT)
              ├── /api/* ──► handler ──► Postgres
              ├── /health ──► health check
              └── /* ──► static SvelteKit build
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
├── Taskfile.yml                     # Task runner — all dev commands
├── docker-compose.yml               # Local Postgres
├── Dockerfile                       # Multi-stage production build
├── sqlc.yml                         # SQL code generation config
├── .air.toml                        # Go hot-reload config
└── .github/workflows/ci.yml        # Lint, test, Docker build
```

## Development Setup

### Prerequisites

- **Go 1.25+** — backend
- **Bun** — frontend package manager and runtime
- **Docker** — local PostgreSQL
- **Task** — task runner ([taskfile.dev](https://taskfile.dev))
- **Air** — Go hot-reload (installed via `go install`)
- **sqlc** — SQL code generation (installed via `go install`)

### Getting Started

```bash
# 1. Clone and enter the repo
git clone <repo-url> && cd motophoto

# 2. Start PostgreSQL
docker compose up -d db

# 3. Install frontend dependencies
cd web && bun install && cd ..

# 4. Copy environment config
cp .env.example .env

# 5. Start both dev servers
task dev
```

This runs Air (Go hot-reload on :8080) and Vite (SvelteKit on :5173) concurrently.

### Task Commands

All commands are defined in `Taskfile.yml`:

| Command | What it does |
|---------|-------------|
| `task dev` | Start both backend (Air) and frontend (Vite) dev servers |
| `task build` | Full production build (frontend + backend) |
| `task build-backend` | Compile Go binary |
| `task build-frontend` | Build SvelteKit |
| `task check` | `go vet` + `svelte-check` |
| `task lint` | `golangci-lint` + `eslint` |
| `task test` | Run Go tests |
| `task generate` | Run sqlc code generation |
| `task docker-build` | Build Docker image |

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | (see `.env.example`) | PostgreSQL connection string |
| `PORT` | `8080` | Go server listen port |

## Database

PostgreSQL 17 runs locally via docker-compose on port **57512** (mapped from container's 5432).

Connection string: `postgres://motophoto:motophoto@localhost:57512/motophoto`

The database layer (pgx pool + sqlc generated queries) exists but is **not yet wired into the API handlers** — the current endpoints return hardcoded demo data. See [GO.md](GO.md) for the sqlc workflow.

## Deployment

### Docker Build

The `Dockerfile` uses a 3-stage build:

1. **Node stage** — installs Bun, builds SvelteKit (`bun run build`)
2. **Go stage** — compiles the Go binary, compresses with UPX
3. **Runtime stage** — Alpine with the compiled binary + frontend build output

### CI/CD

GitHub Actions (`.github/workflows/ci.yml`) runs on every push:

1. **lint-and-test** — Postgres service container, golangci-lint, Go tests, svelte-check, ESLint
2. **docker** (master only) — builds and pushes to GitHub Container Registry (GHCR)

### Railway

Production deploys to Railway, which reads the `Dockerfile` directly. The `PORT` env var is set by Railway.
