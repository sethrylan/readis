// Package data provides Redis data access functionality.
package data

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sethrylan/readis/internal/util"
)

// Data is a wrapper around redis clients; standalone or cluster.
type Data struct {
	cluster bool

	rc *redis.Client
	cc *redis.ClusterClient
}

// Key represents a Redis key
type Key struct {
	Name     string        // the key name
	Datatype string        // Hash, String, Set, etc; https://redis.io/commands/type/
	Size     uint64        // in bytes
	TTL      time.Duration // or -1, if no TTL. Note, in some rare cases, this can be -2.
}

// NewData creates a new Data object for interacting with Redis.
func NewData(uri string, cluster bool) (*Data, error) {
	uri, err := util.NormalizeURI(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid URI: %w", err)
	}

	if cluster {
		clusterOpts, clusterErr := redis.ParseClusterURL(uri)
		if clusterErr != nil {
			return nil, fmt.Errorf("invalid cluster URL: %w", clusterErr)
		}
		return &Data{
			cluster: true,
			cc:      redis.NewClusterClient(clusterOpts),
		}, nil
	}

	options, err := redis.ParseURL(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	return &Data{
		cluster: false,
		rc:      redis.NewClient(options),
	}, nil
}

// URI returns the Redis server address.
func (d *Data) URI() string {
	if d.cluster {
		return d.cc.Options().Addrs[0]
	}
	return d.rc.Options().Addr
}

// Close closes the Redis connection.
func (d *Data) Close() error {
	return d.client().Close()
}

// TotalKeys returns the total number of keys in the Redis database.
func (d *Data) TotalKeys(ctx context.Context) int64 {
	return d.client().DBSize(ctx).Val()
}

// client is a helper function to get the redis client depending on mode (standalone, cluster, etc)
func (d *Data) client() redis.UniversalClient {
	if d.cluster {
		return d.cc
	}
	return d.rc
}

// ScanAsync scans Redis keys asynchronously and returns results via a channel.
func (d *Data) ScanAsync(ctx context.Context, s *Scan) <-chan *Key {
	util.Debug("scan: ", s.pattern, " ", strconv.Itoa(s.pageSize))
	s.scanning.Store(true)
	ch := make(chan *Key)

	go func() {
		defer func() {
			// Close the channel to signal that we're done
			s.scanning.Store(false)
			close(ch)
		}()
		var cmds []redis.Cmder
		var err error

		if strings.Contains(s.pattern, "*") {
			if d.cluster {
				var mu sync.Mutex
				err = d.cc.ForEachMaster(ctx, func(ctx context.Context, rc *redis.Client) error {
					mu.Lock()
					defer mu.Unlock()
					shardCmds, shardErr := s.PipelinedCmds(ctx, rc)
					if shardErr != nil {
						return shardErr
					}
					cmds = append(cmds, shardCmds...)
					return nil
				})
			} else {
				rc := d.rc
				cmds, err = s.PipelinedCmds(ctx, rc)
			}
		} else {
			cmds, err = d.client().Pipelined(ctx, func(pipe redis.Pipeliner) error {
				pipe.TTL(ctx, s.pattern)
				pipe.Type(ctx, s.pattern)
				pipe.MemoryUsage(ctx, s.pattern)
				return nil
			})
		}

		if errors.Is(err, redis.Nil) {
			return
		}
		if err != nil {
			select {
			case ch <- &Key{
				Name:     err.Error(),
				Datatype: "error",
				TTL:      -1,
			}:
			case <-ctx.Done():
			}
			return
		}

		keys := make(map[string]*Key)

		for _, cmd := range cmds {
			key, ok := cmd.Args()[1].(string)
			if !ok {
				continue
			}
			if key == "usage" {
				key, ok = cmd.Args()[2].(string)
				if !ok {
					continue
				}
			}

			if _, ok := keys[key]; !ok {
				keys[key] = &Key{Name: key}
			}

			switch c := cmd.(type) {
			case *redis.DurationCmd:
				keys[key].TTL = c.Val()
			case *redis.StatusCmd:
				keys[key].Datatype = c.Val()
			case *redis.IntCmd:
				val := c.Val()
				if val >= 0 {
					keys[key].Size = uint64(val) // #nosec G115 -- val is checked to be >= 0
				}
			default:
				util.Debug("unknown command type: ", fmt.Sprintf("%T", cmd))
				continue
			}
		}

		for _, key := range keys {
			util.DebugDelay(0.50) // inject delay for testing
			select {
			case ch <- key:
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch
}

// Fetch retrieves the value of a key from Redis and returns it as markdown.
func (d *Data) Fetch(ctx context.Context, key Key) (string, error) {
	c := d.client()

	switch key.Datatype {
	case "string":
		r, err := c.Get(ctx, key.Name).Result()
		if err == nil {
			return fmt.Sprintf("```%s```", r), nil
		}
		return "", err
	case "list":
		var sb strings.Builder
		for _, v := range c.LRange(ctx, key.Name, 0, -1).Val() {
			fmt.Fprintf(&sb, "- `%v`\n", v)
		}
		return sb.String(), nil
	case "set":
		var sb strings.Builder
		for _, v := range c.SMembers(ctx, key.Name).Val() {
			fmt.Fprintf(&sb, "- `%v`\n", v)
		}
		return sb.String(), nil
	case "zset":
		var sb strings.Builder
		sb.WriteString("| score | value |\n| --- | --- |\n")
		for _, z := range c.ZRangeWithScores(ctx, key.Name, 0, -1).Val() {
			fmt.Fprintf(&sb, "| %f | `%v` |\n", z.Score, z.Member)
		}
		return sb.String(), nil
	case "hash":
		hash := c.HGetAll(ctx, key.Name).Val()

		fields := make([]string, 0, len(hash))
		for f := range hash {
			fields = append(fields, f)
		}
		sort.Strings(fields)

		var sb strings.Builder
		sb.WriteString("| field | value |\n| --- | --- |\n")
		for _, f := range fields {
			fmt.Fprintf(&sb, "| %s | %s |\n", f, hash[f])
		}
		return sb.String(), nil
	default:
		return "Unknown data type: " + key.Datatype, nil
	}
}
