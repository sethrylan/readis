package main

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	"github.com/redis/go-redis/v9"
)

type Data struct {
	opts    *redis.UniversalOptions
	cluster bool

	rc *redis.Client
	cc *redis.ClusterClient
}

type Scan struct {
	pageSize int
	pattern  string
	iters    map[string]*redis.ScanIterator
	scanning bool
	// keys     []*Key
}

func (d *Data) TotalKeys(ctx context.Context) int64 {
	if d.cluster {
		return d.cc.DBSize(context.Background()).Val()
	}

	return d.rc.DBSize(context.Background()).Val()
}

// func (d *Data) ResetScan() {
// 	d.pageSize = 0
// 	d.pattern = ""
// 	d.iters = make(map[string]*redis.ScanIterator)
// }

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

func (d *Data) Close() {
	if d.cluster {
		d.cc.Close()
	} else {
		d.rc.Close()
	}
}

func (d *Data) NewScan(pattern string, pageSize int) *Scan {
	debug("new scan: ", pattern, fmt.Sprintf("%d", pageSize))
	// d.ResetScan()
	s := Scan{}
	s.pattern = pattern
	s.pageSize = pageSize
	s.iters = make(map[string]*redis.ScanIterator)
	// s.keys = make([]*Key, 0)

	return &s
}

func (d *Data) scanAsync(s *Scan) (<-chan *Key, context.Context, context.CancelFunc) {
	debug("scan: ", s.pattern, " ", fmt.Sprintf("%d", s.pageSize))

	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan *Key)
	go func() {
		s.scanning = true
		defer func() {
			// Close the channel to signal that we're done
			close(ch)
			s.scanning = false
		}()
		var cmds []redis.Cmder
		var err error
		var numFound int

		// ... (same as in the scan function)

		if d.cluster {
			err = d.cc.ForEachMaster(ctx, func(ctx context.Context, rc *redis.Client) error {
				if s.iters[rc.Options().Addr] == nil {
					s.iters[rc.Options().Addr] = rc.Scan(ctx, 0, s.pattern, int64(s.pageSize)).Iterator()
				}
				iter := s.iters[rc.Options().Addr]
				shardCmds, err := rc.Pipelined(ctx, func(pipe redis.Pipeliner) error {
					for iter.Next(ctx) && numFound < s.pageSize {
						numFound++
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
			if s.iters[rc.Options().Addr] == nil {
				debug("new iterator", rc.Options().Addr)
				s.iters[rc.Options().Addr] = rc.Scan(ctx, 0, s.pattern, int64(s.pageSize)).Iterator()
			}
			iter := s.iters[rc.Options().Addr]

			cmds, err = rc.Pipelined(ctx, func(pipe redis.Pipeliner) error {
				for iter.Next(ctx) && numFound < s.pageSize {
					numFound++
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
				keys[key].datatype = c.Val()
			case *redis.IntCmd:
				keys[key].size = uint64(c.Val())
			default:
				panic("unknown type")
			}
		}

		for _, key := range keys {
			// Send the key to the channel
			ch <- key
		}
	}()

	return ch, ctx, cancel
}

func (d *Data) scan(ctx context.Context, s *Scan) map[string]*Key {
	var cmds []redis.Cmder
	var err error
	var numFound int

	debug("scan", s.pattern, fmt.Sprintf("%d", s.pageSize))

	if d.cluster {
		err = d.cc.ForEachMaster(ctx, func(ctx context.Context, rc *redis.Client) error {
			if s.iters[rc.Options().Addr] == nil {
				s.iters[rc.Options().Addr] = rc.Scan(ctx, 0, s.pattern, int64(s.pageSize)).Iterator()
			}
			iter := s.iters[rc.Options().Addr]
			shardCmds, err := rc.Pipelined(ctx, func(pipe redis.Pipeliner) error {
				for iter.Next(ctx) && numFound < s.pageSize {
					numFound++
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
		if s.iters[rc.Options().Addr] == nil {
			debug("new iterator", rc.Options().Addr)
			s.iters[rc.Options().Addr] = rc.Scan(ctx, 0, s.pattern, int64(s.pageSize)).Iterator()
		}
		iter := s.iters[rc.Options().Addr]

		cmds, err = rc.Pipelined(ctx, func(pipe redis.Pipeliner) error {
			for iter.Next(ctx) && numFound < s.pageSize {
				numFound++
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
			keys[key].datatype = c.Val()
		case *redis.IntCmd:
			keys[key].size = uint64(c.Val())
		default:
			panic("unknown type")
		}
	}

	return keys
}

func (d *Data) ScanMore(ctx context.Context, s *Scan) []list.Item {
	var items []list.Item
	for _, v := range d.scan(ctx, s) {
		items = append(items, *v)
	}
	return items
}

func (d *Data) Fetch(key Key) string {
	var ctx = context.Background()
	var uc redis.UniversalClient
	if d.cluster {
		uc = d.cc
	} else {
		uc = d.rc
	}

	switch key.datatype {
	case "string":
		r, err := uc.Get(ctx, key.name).Result()
		if err == nil {
			return fmt.Sprintf("```%s```", r)
		}
	case "list":
		markdown := ""
		for _, v := range uc.LRange(ctx, key.name, 0, -1).Val() {
			markdown += fmt.Sprintf("- `%v`\n", v)
		}
		return markdown
	case "set":
		markdown := ""
		for _, v := range uc.SMembers(ctx, key.name).Val() {
			markdown += fmt.Sprintf("- `%v`\n", v)
		}
		return markdown
	case "zset":
		markdown := "| score | value |\n| --- | --- |\n"
		for _, z := range uc.ZRangeWithScores(ctx, key.name, 0, -1).Val() {
			markdown += fmt.Sprintf("| %f | `%v` |\n", z.Score, z.Member)
		}
		return markdown
	case "hash":
		markdown := "| field | value |\n| --- | --- |\n"
		for k, v := range uc.HGetAll(ctx, key.name).Val() {
			markdown += fmt.Sprintf("| %s | %s |\n", k, v)
		}
		return markdown
	default:
		return `
| Name        | Price | Notes                           |
| ---         | ---   | ---                             |
| Tsukemono   | $2    | Just an appetizer               |
| Tomato Soup | $4    | Made with San Marzano tomatoes  |
| Okonomiyaki | $4    | Takes a few minutes to make     |
| Curry       | $3    | We can add squash if you’d like |`

	}

	return "could not get value for " + key.datatype
}
