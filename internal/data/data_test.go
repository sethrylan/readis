package data //nolint:testpackage // white-box testing of internal package

import (
	"context"
	"strconv"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	redisTestContainers "github.com/testcontainers/testcontainers-go/modules/redis"
)

func TestNewData(t *testing.T) {
	ctx := context.Background()
	c, d, redisContainer := setup(t)

	defer func() {
		err := c.Close()
		if err != nil {
			panic(err)
		}
		err = redisContainer.Terminate(ctx)
		if err != nil {
			panic(err)
		}
	}()

	// populate test data
	for i := range 1000 {
		_, err := c.Set(ctx, "testkey:"+strconv.Itoa(i), "testvalue", 0).Result()
		require.NoError(t, err)
	}

	assert.Equal(t, int64(1000), d.TotalKeys(ctx))
	err := d.Close()
	require.NoError(t, err)
}

func setup(t *testing.T) (*redis.Client, *Data, testcontainers.Container) {
	t.Helper()
	ctx := context.Background()
	redisContainer, _ := redisTestContainers.Run(ctx,
		"docker.io/redis:7.2",
		redisTestContainers.WithLogLevel(redisTestContainers.LogLevelVerbose),
	)

	connStr, err := redisContainer.ConnectionString(ctx)
	require.NoError(t, err)
	opts, err := redis.ParseURL(connStr)
	require.NoError(t, err)

	c := redis.NewClient(opts)
	d, err := NewData(connStr, false)
	require.NoError(t, err)

	return c, d, redisContainer
}
