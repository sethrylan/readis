package data

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/sethrylan/readis/internal/util"
	"github.com/redis/go-redis/v9"
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
func NewData(uri string, cluster bool) *Data {
	uri = util.NormalizeURI(uri)

	if cluster {
		options := util.PanicOnError(redis.ParseClusterURL(uri))
		return &Data{
			cluster: true,
			cc:      redis.NewClusterClient(options),
		}
	}

	options := util.PanicOnError(redis.ParseURL(uri))
	return &Data{
		cluster: false,
		rc:      redis.NewClient(options),
	}
}

func (d *Data) URI() string {
	if d.cluster {
		return d.cc.Options().Addrs[0]
	}
	return d.rc.Options().Addr
}

func (d *Data) Close() error {
	return d.client().Close()
}

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

func (d *Data) ScanAsync(ctx context.Context, s *Scan) <-chan *Key {
	util.Debug("scan: ", s.pattern, " ", fmt.Sprintf("%d", s.pageSize))
	s.scanning = true
	ch := make(chan *Key)

	go func() {
		defer func() {
			// Close the channel to signal that we're done
			s.scanning = false
			close(ch)
		}()
		var cmds []redis.Cmder
		var err error

		if strings.Contains(s.pattern, "*") {
			if d.cluster {
				err = d.cc.ForEachMaster(ctx, func(ctx context.Context, rc *redis.Client) error {
					shardCmds, err := s.PipelinedCmds(ctx, rc)
					if err != nil {
						return err
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
			ch <- &Key{
				Name:     err.Error(),
				Datatype: "error",
				TTL:      -1,
			}
			return
		}

		keys := make(map[string]*Key)

		for _, cmd := range cmds {
			key := cmd.Args()[1].(string)
			if key == "usage" {
				key = cmd.Args()[2].(string)
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
				keys[key].Size = uint64(c.Val())
			default:
				panic("unknown type")
			}
		}

		for _, key := range keys {
			util.DebugDelay(0.50) // inject delay for testing
			ch <- key             // Send the key to the channel
		}
	}()

	return ch
}

func (d *Data) Fetch(ctx context.Context, key Key) string {
	c := d.client()

	switch key.Datatype {
	case "string":
		r, err := c.Get(ctx, key.Name).Result()
		if err == nil {
			return fmt.Sprintf("```%s```", r)
		}
	case "list":
		markdown := ""
		for _, v := range c.LRange(ctx, key.Name, 0, -1).Val() {
			markdown += fmt.Sprintf("- `%v`\n", v)
		}
		return markdown
	case "set":
		markdown := ""
		for _, v := range c.SMembers(ctx, key.Name).Val() {
			markdown += fmt.Sprintf("- `%v`\n", v)
		}
		return markdown
	case "zset":
		markdown := "| score | value |\n| --- | --- |\n"
		for _, z := range c.ZRangeWithScores(ctx, key.Name, 0, -1).Val() {
			markdown += fmt.Sprintf("| %f | `%v` |\n", z.Score, z.Member)
		}
		return markdown
	case "hash":
		hash := c.HGetAll(ctx, key.Name).Val()

		fields := make([]string, 0)
		for f := range hash {
			fields = append(fields, f)
		}
		sort.Strings(fields)

		markdown := "| field | value |\n| --- | --- |\n"
		for _, f := range fields {
			markdown += fmt.Sprintf("| %s | %s |\n", f, hash[f])
		}
		return markdown
	default:
		return "Unknown data type: " + key.Datatype
	}

	return "could not get value for " + key.Datatype
}
