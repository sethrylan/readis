# readis

Readis is a TUI [Redis](https://redis.io/) browser, built with ❤️ and [Charm](https://charm.sh/), inspired by [Redis Insight](https://redislabs.com/redis-enterprise/redis-insight/) [and](https://github.com/snmaynard/redis-audit) [other](https://github.com/antirez/redis-sampler) [tools](https://github.com/gamenet/redis-memory-analyzer).


## Installing

```sh
go install github.com/github/readis@latest
```

## Using

```sh
readis --help

readis // defaults to localhost:6379

readis -c rediss://mycluster.example.com:10000
```
