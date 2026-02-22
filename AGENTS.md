# AGENTS.md

## Project Overview

Readis is a TUI Redis browser written in Go, built with [Charm](https://charm.sh/) (Bubble Tea, Lip Gloss, Glamour) and inspired by RedisInsight.

- **Language:** Go (version specified in `go.mod`)
- **Entry point:** `cmd/readis/main.go`
- **Structure:** `internal/data` (Redis data layer), `internal/ui` (TUI components), `internal/util` (utilities)

## Building and Running

```sh
go build ./cmd/readis
go run cmd/readis/main.go
```

## Testing

```sh
go test -race ./...
```

Tests use [testcontainers-go](https://github.com/testcontainers/testcontainers-go) with a Redis module, so Docker must be available to run the full test suite.

## Linting

```sh
golangci-lint run
```

The project uses an extensive `golangci-lint` configuration (`.golangci.yml`) with nearly all linters enabled. Key points:

- Linters are set to `default: all` with specific noisy linters disabled.
- Formatters: `gofmt` and `goimports` are enforced.
- Test files have relaxed rules for `gosec`, `errcheck`, and dot-imports.
- Run `go mod tidy` before submitting changes.

## Code Style and Conventions

- Follow standard Go conventions and idioms.
- Dot imports are allowed in test files (for testify).
- US English spelling is enforced by the `misspell` linter.

## CI/CD

CI runs on every push and PR to `main` via GitHub Actions (`.github/workflows/ci.yml`):

1. **Lint** — runs `golangci-lint`
2. **Test** — runs `go test -v -race -coverprofile=coverage.out ./...`

### Additional Workflows

- **Conventional Commits** — PR titles must follow the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) specification.
- **Dependabot Auto-merge** — Patch and minor dependency updates from Dependabot are auto-approved and squash-merged.
- **Release** — Triggered by pushing a `v*.*.*` tag. Uses [GoReleaser](https://goreleaser.com/) to build and publish binaries for Linux and macOS.

## PR Guidelines

- PR titles must follow the Conventional Commits format (e.g., `feat: add feature`, `fix: resolve bug`).
- Ensure `go mod tidy`, `golangci-lint run`, and `go test -race ./...` all pass before submitting.
