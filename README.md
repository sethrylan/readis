# readis

[![Latest Release](https://img.shields.io/github/release/sethrylan/readis.svg)](https://github.com/sethrylan/readis/releases)
[![Build Status](https://github.com/sethrylan/readis/workflows/CI/badge.svg)](https://github.com/sethrylan/readis/actions)

A TUI [Redis](https://redis.io/) browser, built with [Charm](https://charm.sh/) and inspired by [RedisInsight](https://redislabs.com/redis-enterprise/redis-insight/).

![demo video](https://raw.githubusercontent.com/sethrylan/readis/main/docs/demo.gif)

## Installing

Download the [latest release](https://github.com/sethrylan/readis/releases)

or

```sh
➜ go install github.com/sethrylan/readis@main
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
