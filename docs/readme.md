# Demo

The `demo.gif` in the project README is generated automatically.

## How It Works

- **[demo.tape](demo.tape)** — A [VHS](https://github.com/charmbracelet/vhs) script that records a terminal session of `readis` as a GIF.
- **[demo.txt](demo.txt)** — Redis commands used to prepopulate the server with interesting demo data. One command per line; blank lines and lines starting with `#` are ignored.

## CI Workflow

The [Demo workflow](../.github/workflows/demo.yml) runs automatically on PRs to `main` when relevant files change (`docs/demo.tape`, `docs/demo.txt`, `cmd/**`, `internal/**`, `go.mod`, `go.sum`). It builds `readis`, starts Redis, loads the demo data, runs VHS, and commits the updated `demo.gif` back to the PR branch.
