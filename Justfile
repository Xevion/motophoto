# MotoPhoto - Event Photography Marketplace

set dotenv-load := true
set positional-arguments := true

alias c := check
alias d := dev
alias t := test
alias test-cover := cov

default:
    @just --list

# Validate all code (parallel checks)
check *flags:
    tempo check {{flags}}

# Dev server - frontend + backend. Flags: -f(rontend) -b(ackend)
dev *flags:
    -tempo dev {{flags}}

# Run all tests (use -v for verbose, -run=<pattern> to filter)
test *flags:
    go test -race -count=1 {{flags}} ./...
    cd web && bunx vitest run

# Run tests with coverage report and diff against master baseline
cov:
    tempo run cov

# Run linters
lint *flags:
    tempo lint {{flags}}

# Auto-format all code
format *flags:
    tempo fmt {{flags}}

# Build everything for production
build:
    bun run --cwd web build
    go build -o motophoto .

# Run sqlc and tygo code generation
generate:
    sqlc generate
    tygo generate

# Build Docker image
docker-build *flags:
    docker build -t motophoto:latest {{flags}} .

# Run Docker image
docker-run *flags:
    docker run --rm -it --network host {{flags}} motophoto:latest

# Manage local Postgres via docker-compose
db cmd="start":
    #!/usr/bin/env bash
    set -euo pipefail
    if ! command -v docker &>/dev/null; then
        echo "x docker not found -- install Docker first (see README.md)" >&2
        exit 1
    fi
    if ! docker info &>/dev/null; then
        echo "x Docker daemon is not running -- start Docker first" >&2
        exit 1
    fi
    case "{{cmd}}" in
        start)
            docker compose up -d --wait db
            ;;
        reset)
            docker compose up -d --wait db
            docker compose exec db psql -U motophoto -d postgres -c "DROP DATABASE IF EXISTS motophoto"
            docker compose exec db psql -U motophoto -d postgres -c "CREATE DATABASE motophoto"
            echo "Database reset. Restart the dev server to re-run migrations."
            ;;
        rm)
            docker compose down
            ;;
        *)
            echo "Unknown command: {{cmd}}" >&2
            exit 1
            ;;
    esac

# Clean build artifacts
clean:
    rm -rf motophoto web/build web/.svelte-kit tmp/
