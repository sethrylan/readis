# Demo

The `demo.gif` in the project README is generated automatically.

## How It Works

- **[demo-setup.sh](demo-setup.sh)** — Builds `readis`, starts a Redis container, and populates it with demo data. Run from the repo root:
  ```sh
  docs/demo-setup.sh
  ```
- **[demo.tape](demo.tape)** — A [VHS](https://github.com/charmbracelet/vhs) script that records a terminal session of `readis` as a GIF.

## CI Workflow

The [Demo workflow](../.github/workflows/demo.yml) runs automatically on PRs to `main` when relevant files change (`docs/demo.tape`, `docs/demo-setup.sh`, `cmd/**`, `internal/**`, `go.mod`, `go.sum`). It calls `demo-setup.sh` to build `readis` and prepare the Redis environment, then runs VHS to generate the GIF and commits the updated `demo.gif` back to the PR branch.
