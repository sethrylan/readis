package main

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

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
