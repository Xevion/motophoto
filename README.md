# MotoPhoto

Event photography marketplace — photographers upload action sports photos, customers browse and purchase them.

**Tech stack**: Go (Chi, pgx, sqlc) + SvelteKit 2 (Svelte 5, Bun) + PostgreSQL

## Quickstart

### Prerequisites

All tools are defined in `.mise.toml` and managed by [mise](https://mise.jdx.dev):

```bash
# Install mise (see https://mise.jdx.dev/getting-started.html), then:
mise trust && mise install
```

This puts `go`, `bun`, `just`, `air`, `sqlc`, and other tools on your PATH automatically.

<details>
<summary>Manual install (without mise)</summary>

Install each tool individually:

- [Go 1.26+](https://go.dev)
- [Bun](https://bun.sh)
- [just](https://just.systems)
- [Docker](https://docs.docker.com/engine/install/) — for local PostgreSQL

Go-based tools:

```bash
go install github.com/air-verse/air@latest
go install github.com/gzuidhof/tygo@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```
</details>

### Getting started

```bash
git clone <repo-url> && cd motophoto
docker compose up -d db        # Start PostgreSQL
cd web && bun install && cd ..  # Install frontend dependencies
cp .env.example .env            # Configure environment
just dev                        # Start dev servers (Go :3001 + SvelteKit :5173)
```

### Platform notes

- **Linux** — develop natively. Install [Docker Engine](https://docs.docker.com/engine/install/) for local Postgres.
- **Windows** — develop inside [WSL 2](https://learn.microsoft.com/en-us/windows/wsl/install). Docker Engine installs directly inside WSL.

## Commands

All commands are defined in `Justfile`:

| Command | What it does |
|---------|-------------|
| `just dev` | Start both backend (Air) and frontend (Vite) dev servers |
| `just dev -f` | Frontend only |
| `just dev -b` | Backend only |
| `just build` | Full production build (frontend + backend) |
| `just check` | Parallel: `go vet` + `go build` + `go test` + `svelte-check` + `eslint` |
| `just check --fix` | Auto-format then verify |
| `just lint` | `eslint` + `go vet` |
| `just test` | Run Go tests |
| `just format` | `gofmt` + `eslint --fix` |
| `just generate` | Run sqlc code generation |
| `just docker-build` | Build Docker image |
| `just db` | Start local Postgres |
| `just db reset` | Drop and recreate database |

## Documentation

| Guide | Covers |
|-------|--------|
| [Architecture](docs/ARCHITECTURE.md) | System design, request flow, project structure, deployment |
| [Go Backend](docs/GO.md) | Router, handlers, sqlc, error handling, testing |
| [SvelteKit Frontend](docs/SVELTE.md) | Routing, data fetching, Svelte 5, configuration |
| [Style Guide](docs/STYLE.md) | Naming, vocabulary, comments, API design, logging |
