# PRs Welcome!

## Testing your code
1. Run `go test -race ./...`
1. Validate your code by running `go run cmd/readis/main.go` in the root directory.
1. Try with a local Redis instance by running
    - `docker run --rm --name redis -p 6379:6379 redis --enable-debug-command yes`
1. Or with a more configured redis instance
```sh
cat > redis.conf <<EOF
requirepass foobared
enable-debug-command yes
EOF

docker run --rm -v .:/usr/local/etc/redis --name redis -p 6379:6379 redis redis-server /usr/local/etc/redis/redis.conf
```

Use the undocumented `DEBUG POPULATE` command to populate the database with some data.

```sh
docker exec redis redis-cli DEBUG POPULATE 1000 testkeys 4096
```


### Debug options
- The `--debug` flag will print debug logs to the debug.log file.
- The `DEBUG_DELAY` env var will add additional delay to database commands. Useful for simulating network delay. E.g., `DEBUG_DELAY=100 go run cmd/readis/main.go` to add an average of 100ms of randomized delay

## Reviewing your code
1. Run `go mod tidy`
1. Run `golangci-lint run`
1. Update the demo video, if needed.

## Releasing a new version
1. Go to https://github.com/sethrylan/readis/releases to create a new release by clicking "Draft a new release" with a new tag.
