package exec

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

// export REDIS_URI=$(vault-secret --application actions-redis -e lab  | jq -r '.ACTIONS_REDIS_URIS | split(" ")[0]')
func main() {
	find_keys_without_ttl()
}

func find_keys_without_ttl() {
	opts, err := redis.ParseClusterURL(os.Getenv("REDIS_URI"))
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	cc := redis.NewClusterClient(opts)
	var cmds []redis.Cmder
	var ctx = context.Background()

	fmt.Println("total:", cc.DBSize(ctx).Val())

	err = cc.ForEachMaster(ctx, func(ctx context.Context, rc *redis.Client) error {
		fmt.Println("master:", rc.Options().Addr)
		scan := rc.Scan(ctx, 0, "*", 1000)
		iter := scan.Iterator()
		page, cursor := scan.Val()

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

		fmt.Println("cursor:", cursor)
		fmt.Println("pagesize:", len(page))

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
