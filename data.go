package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/redis/go-redis/v9"
)

type Data struct {
	opts    *redis.UniversalOptions
	cluster bool

	rc *redis.Client
	cc *redis.ClusterClient

	cursor uint64
}

func (d *Data) ResetCursor() {
	d.cursor = 0
}

func NewData() *Data {
	d := &Data{}
	uri, found := os.LookupEnv("REDIS_URI")
	if !found {
		d.cluster = false
		d.opts = &redis.UniversalOptions{
			Addrs: []string{"localhost:6379"},
		}
		d.rc = redis.NewClient(d.opts.Simple())
	} else {
		options := panicOnError(redis.ParseClusterURL(uri))
		d.cluster = true
		d.opts = &redis.UniversalOptions{
			Addrs:           options.Addrs,
			Username:        options.Username,
			Password:        options.Password,
			TLSConfig:       options.TLSConfig,
			MaxRetries:      options.MaxRetries,
			MinRetryBackoff: options.MinRetryBackoff,
			MaxRetryBackoff: options.MaxRetryBackoff,
			DialTimeout:     options.DialTimeout,
			ReadTimeout:     options.ReadTimeout,
			WriteTimeout:    options.WriteTimeout,
			PoolSize:        options.PoolSize,
			MinIdleConns:    options.MinIdleConns,
			PoolTimeout:     options.PoolTimeout,
			MaxRedirects:    options.MaxRedirects,
			ReadOnly:        options.ReadOnly,
			RouteByLatency:  options.RouteByLatency,
			RouteRandomly:   options.RouteRandomly,
		}
		d.cc = redis.NewClusterClient(d.opts.Cluster())
	}

	return d
}

func randtype() string {
	types := []string{
		"set",
		"sorted set",
		"hash",
		"string",
		"list",
	}
	n := rand.Int() % len(types)
	return types[n]

}

var allkeys = [...]list.Item{
	Key{name: "Raspberry Pi’s", keyType: randtype(), size: rand.Intn(100), ttl: time.Duration(rand.Intn(100000000000))},
	Key{name: "Nutella", keyType: randtype(), size: 12, ttl: 0},
	Key{name: "Bitter melon", keyType: randtype(), size: 12, ttl: 0},
	Key{name: "Nice socks", keyType: randtype(), size: 12, ttl: 0},
	Key{name: "Eight hours of sleep", keyType: randtype(), size: 12, ttl: 0},
	Key{name: "Cats", keyType: randtype(), size: 12, ttl: 0},
	Key{name: "Plantasia, the album", keyType: randtype(), size: 12, ttl: 0},
	Key{name: "Pour over coffee", keyType: randtype(), size: 12, ttl: 0},
	Key{name: "VR", keyType: randtype(), size: 12, ttl: 0},
	Key{name: "Noguchi Lamps", keyType: "hash", size: 12, ttl: 0},
	Key{name: "Linux", keyType: "hash", size: 12, ttl: 0},
	Key{name: "Business school", keyType: "hash", size: 12, ttl: 0},
	Key{name: "Pottery", keyType: "hash", size: 12, ttl: 0},
	Key{name: "Shampoo", keyType: "hash", size: 12, ttl: 0},
	Key{name: "Table tennis", keyType: "hash", size: 12, ttl: 0},
	Key{name: "Milk crates", keyType: "hash", size: 12, ttl: 0},
	Key{name: "Afternoon tea", keyType: "hash", size: 12, ttl: 0},
	Key{name: "Stickers", keyType: "hash", size: 12, ttl: 0},
	Key{name: "20° Weather", keyType: "hash", size: 12, ttl: 0},
	Key{name: "Warm light", keyType: "hash", size: 12, ttl: 0},
	Key{name: "The vernal equinox", keyType: "hash", size: 12, ttl: 0},
	Key{name: "Gaffer’s tape", keyType: "hash", size: 12, ttl: 0},
	Key{name: "Terrycloth", keyType: "hash", size: 12, ttl: 0},
}

func (*Data) ScanMock(n int) (int, int, []list.Item) {
	n = min(n, len(allkeys))
	return n, n * 100, allkeys[:n]
}

// Scan returns the number of keys scanned, the total keys, and the keys found
func (d *Data) Scan(pattern string, cursor int64, count int64) (int64, int64, []list.Item) {
	var cmds []redis.Cmder
	var err error
	var ctx = context.Background()
	var n, total int64

	if d.cluster {
		total = d.cc.DBSize(ctx).Val()
		err = d.cc.ForEachMaster(ctx, func(ctx context.Context, rc *redis.Client) error {
			scan := rc.Scan(ctx, 0, pattern, 1000000)
			page, cursor := scan.Val()
			d.cursor = cursor
			n += int64(len(page))
			iter := scan.Iterator()
			ttlcmds, err := rc.Pipelined(ctx, func(pipe redis.Pipeliner) error {
				for iter.Next(ctx) {
					pipe.TTL(ctx, iter.Val())
				}
				return nil
			})
			if err != nil {
				panic(err)
			}
			cmds = append(cmds, ttlcmds...)
			return iter.Err()
		})
	} else {
		total = d.rc.DBSize(ctx).Val()
		scan := d.rc.Scan(ctx, 0, pattern, 1000000)
		page, cursor := scan.Val()
		d.cursor = cursor
		n += int64(len(page))
		iter := d.rc.Scan(ctx, 0, "*", 1000000).Iterator()
		ttlcmds, err := d.rc.Pipelined(ctx, func(pipe redis.Pipeliner) error {
			for iter.Next(ctx) {
				pipe.TTL(ctx, iter.Val())
			}
			return nil
		})
		if err != nil {
			panic(err)
		}
		cmds = append(cmds, ttlcmds...)
	}

	if err != nil {
		panic(err)
	}

	for _, cmd := range cmds {
		switch v := cmd.(type) {
		case *redis.DurationCmd:
			fmt.Printf("Twice %v is %v\n", v)
		default:
			fmt.Printf("I don't know about type %T!\n", v)
		}

	}

	return n, total, nil

}

//////////////////////////

func find_keys_without_ttl() {
	opts, err := redis.ParseClusterURL(os.Getenv("REDIS_URI"))
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	rc := redis.NewClusterClient(opts)
	var cmds []redis.Cmder
	var ctx = context.Background()

	err = rc.ForEachMaster(ctx, func(ctx context.Context, rc *redis.Client) error {
		iter := rc.Scan(ctx, 0, "*", 1000000).Iterator()
		shardcmds, err := rc.Pipelined(ctx, func(pipe redis.Pipeliner) error {
			for iter.Next(ctx) {
				pipe.TTL(ctx, iter.Val())
			}
			return nil
		})
		if err != nil {
			panic(err)
		}
		cmds = append(cmds, shardcmds...)
		return iter.Err()
	})

	if err != nil {
		panic(err)
	}

	for _, cmd := range cmds {
		if cmd.(*redis.DurationCmd).Val() == -1 {
			fmt.Println(cmd.Args()[1])
		}
	}
}
