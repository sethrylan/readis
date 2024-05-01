# PRs Welcome!

## Testing your code
1. Run `go test -race ./...`
1. Validate your code by running `go run cmd/readis/main.go` in the root directory.
1. Try with a local Redis instance by running
    - `docker run --rm --name redis -p 6379:6379 redis --enable-debug-command yes`
1. Or with a more configured redis instance
```
cat > redis.conf <<EOF
requirepass foobared
enable-debug-command yes
EOF

docker run --rm -v .:/usr/local/etc/redis --name redis -p 6379:6379 redis redis-server /usr/local/etc/redis/redis.conf
```
1. Use the undocumented `DEBUG POPULATE` command to populate the database with some data.
    - `DEBUG POPULATE 1000 testkeys 4096`

### Debug options
- The `--debug` flag will print debug logs to the debug.log file.
- The `DEBUG_DELAY` env var will add additional delay to database commands. Useful for simulating network delay. E.g., `DEBUG_DELAY=100 go run cmd/readis/main.go` to add an average of 100ms of randomized delay

## Reviewing your code
1. Run `go mod tidy`
1. Run `golangci-lint run`
1. Update the demo video, if needed.

## Releasing a new version
1. Once merged, checkout the `main` branch and pull changes.
1. On the `main` branch, make a new tag for the desired version number, by running `git tag -a vX.X.X`, where `vX.X.X` is the version number (like `v0.0.10`).
1. When prompted for a note, write a brief description of the changes in the new version.
1. Run `git push origin vX.X.X` to push the new tag to the remote.
1. Once that is completed, go to https://github.com/sethrylan/readis/releases to create a new release by clicking "Draft a new release".
