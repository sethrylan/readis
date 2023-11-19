package main

import (
	"context"
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

	total   int64
	count   int64
	pattern string
	cursors map[string]uint64
	scanned map[string]uint64
}

// TotalScanned returns an approximate total number of keys scanned so far.
func (d *Data) TotalFound() int64 {
	var n int64
	for _, c := range d.scanned {
		n += int64(c)
	}
	return n
}

func (d *Data) TotalKeys() int64 {
	if d.cluster {
		return d.cc.DBSize(context.Background()).Val()
	} else {
		return d.rc.DBSize(context.Background()).Val()
	}
}

func (d *Data) ResetScan() {
	d.total = 0
	d.count = 0
	d.pattern = ""
	d.cursors = make(map[string]uint64)
	d.scanned = make(map[string]uint64)
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
	Key{name: "Raspberry Pi’s", keyType: randtype(), size: int64(rand.Intn(100)), ttl: time.Duration(rand.Intn(100000000000))},
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
func (d *Data) NewScan(pattern string, count int64) []list.Item {
	var ctx = context.Background()

	d.ResetScan()
	d.pattern = pattern
	d.count = count

	keys := d.scan(ctx)

	var items []list.Item
	for _, key := range keys {
		items = append(items, *key)
	}

	return items
}

func (d *Data) scan(ctx context.Context) map[string]*Key {
	var cmds []redis.Cmder
	var err error

	if d.cluster {
		err = d.cc.ForEachMaster(ctx, func(ctx context.Context, rc *redis.Client) error {
			currCursor := d.cursors[rc.Options().Addr]
			scan := rc.Scan(ctx, currCursor, d.pattern, d.count)
			_, nextCursor := scan.Val()
			d.cursors[rc.Options().Addr] = nextCursor
			iter := scan.Iterator()

			shardCmds, err := rc.Pipelined(ctx, func(pipe redis.Pipeliner) error {
				for iter.Next(ctx) {
					d.scanned[rc.Options().Addr]++
					pipe.TTL(ctx, iter.Val())
					pipe.Type(ctx, iter.Val())
					pipe.MemoryUsage(ctx, iter.Val())
				}
				return nil
			})
			if err != nil {
				panic(err)
			}
			cmds = append(cmds, shardCmds...)
			return iter.Err()
		})
	} else {
		rc := d.rc

		currCursor := d.cursors[rc.Options().Addr]
		scan := rc.Scan(ctx, currCursor, d.pattern, d.count)
		_, nextCursor := scan.Val()
		d.cursors[rc.Options().Addr] = nextCursor

		iter := scan.Iterator()

		cmds, err = rc.Pipelined(ctx, func(pipe redis.Pipeliner) error {
			for iter.Next(ctx) {
				d.scanned[rc.Options().Addr]++
				pipe.TTL(ctx, iter.Val())
				pipe.Type(ctx, iter.Val())
				pipe.MemoryUsage(ctx, iter.Val())
			}
			return nil
		})
	}

	if err != nil {
		panic(err)
	}

	keys := make(map[string]*Key)

	for _, cmd := range cmds {
		key := cmd.Args()[1].(string)
		if key == "usage" {
			key = cmd.Args()[2].(string)
		}

		if _, ok := keys[key]; !ok {
			keys[key] = &Key{name: key}
		}

		switch c := cmd.(type) {
		case *redis.DurationCmd:
			keys[key].ttl = c.Val()
		case *redis.StatusCmd:
			keys[key].keyType = c.Val()
		case *redis.IntCmd:
			keys[key].size = c.Val()
		default:
			panic("unknown type")
		}
	}

	return keys
}

func (d *Data) ScanMore() []list.Item {
	var items []list.Item
	for _, key := range d.scan(context.Background()) {
		items = append(items, *key)
	}

	return items
}
