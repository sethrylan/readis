# readis

Readis is a TUI Redis browser, built with [Charm](https://charm.sh/) libraries, inspired by [Redis Insight](https://redislabs.com/redis-enterprise/redis-insight/) [and](https://github.com/snmaynard/redis-audit) [other](https://github.com/antirez/redis-sampler) [tools](https://github.com/gamenet/redis-memory-analyzer).


# Running readis

```bash
go install github.com/sethrylan/readis@latest
readis
```

# Notes

docker run --name redis -p 6379:6379 redis --enable-debug-command yes

DEBUG POPULATE 1000 test 40

HSET myhash field1 "Hello"
HSET myhash field2 "Hi" field3 "World"
