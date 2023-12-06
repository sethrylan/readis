# readis

Readis is a TUI Redis browser, built with [Charm](https://charm.sh/) libraries, inspired by [Redis Insight](https://redislabs.com/redis-enterprise/redis-insight/) [and](https://github.com/snmaynard/redis-audit) [other](https://github.com/antirez/redis-sampler) [tools](https://github.com/gamenet/redis-memory-analyzer).


# Running readis

```bash
go install github.com/sethrylan/readis@latest

readis --help

readis // defaults to localhost:6379

readis -c rediss://mycluster.example.com:10000
```
