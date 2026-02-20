# Setup

All project tools are defined in `.mise.toml` and managed by [mise](https://mise.jdx.dev).

## First-time setup

1. **Install mise** — see [mise: Getting Started](https://mise.jdx.dev/getting-started.html)

2. **Add the shell hook** to the end of your shell rc file (`~/.bashrc`, `~/.zshrc`, etc.):
   ```bash
   eval "$(mise activate bash)"   # or zsh, fish
   ```
   Restart your shell (or `source` the rc file) afterward.

3. **Trust and install** from the project root:
   ```bash
   mise trust
   mise install
   ```

`mise trust` marks the project config as safe to use. `mise install` downloads and installs all tools. After this, `go`, `bun`, `just`, and everything else is on your PATH automatically when you're in the project directory.

## Tools

| Tool | Purpose | Required |
|------|---------|----------|
| [go](https://go.dev) | Backend language | Yes |
| [bun](https://bun.sh) | Frontend runtime, script runner | Yes |
| [just](https://just.systems) | Task runner | Yes |
| [air](https://github.com/air-verse/air) | Go hot-reload dev server | For `just dev` |
| [goimports](https://pkg.go.dev/golang.org/x/tools/cmd/goimports) | Go import formatting | No — skipped if missing |
| [golangci-lint](https://golangci-lint.run) | Go linter | No — skipped if missing |
| [tygo](https://github.com/gzuidhof/tygo) | Go → TypeScript type generation | No — skipped if missing |
| [sqlc](https://sqlc.dev) | SQL → Go code generation | For `just generate` |
| [docker](https://docs.docker.com/engine/install/) | Local Postgres | For `just db` (Linux only) |

Biome (frontend formatter) is installed as an npm dev dependency in `web/` — no separate install needed.

## Platform notes

### Linux

Develop natively. Install [Docker Engine](https://docs.docker.com/engine/install/) for local Postgres (`just db`). Docker Desktop is not needed.

### Windows

Develop inside [WSL 2](https://learn.microsoft.com/en-us/windows/wsl/install). Clone the repo, install mise, and run everything from within the WSL shell. Docker Engine installs directly inside WSL — Docker Desktop is not required.

Do not develop from Windows natively; the toolchain assumes a Unix environment.

## Manual install (without mise)

If you prefer not to use mise, install each tool individually — see the links above. Go-based tools can also be installed with `go install`:

```bash
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/air-verse/air@latest
go install github.com/gzuidhof/tygo@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

Docker must be installed separately — mise does not manage it.
