# readis

Readis is a TUI [Redis](https://redis.io/) browser, built with ❤️ and [Charm](https://charm.sh/), inspired by [RedisInsight](https://redislabs.com/redis-enterprise/redis-insight/) [and](https://github.com/snmaynard/redis-audit) [other](https://github.com/antirez/redis-sampler) [tools](https://github.com/gamenet/redis-memory-analyzer).


## Installing

```sh
➜ go install github.com/github/readis@main
```
or

```sh
➜ gh repo clone github/readis && \
  cd readis && \
  go build .
```

## Using

```sh
# print help and options
➜ readis --help

# try a local redis instance (the default)
➜ readis

# try a cluster URI
➜ readis -c rediss://user:$pass@mycluster.example.com:10000
```
