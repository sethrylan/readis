# readis

A TUI [Redis](https://redis.io/) browser, built with [Charm](https://charm.sh/) and inspired by [RedisInsight](https://redislabs.com/redis-enterprise/redis-insight/).

![demo video](./docs/demo.gif)

## Installing

```sh
➜ go install github.com/github/readis@main
```
or

```sh
➜ gh repo clone github/readis && cd readis
➜ go build .
```

## Using

```sh
# print help and options
➜ readis --help

# connect to a local redis instance
➜ readis

# connect to a cluster
➜ readis -c rediss://user:$pass@mycluster.example.com:10000
```
