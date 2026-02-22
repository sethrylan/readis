package data //nolint:testpackage // white-box testing of internal package

import (
	"context"
	"os"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	redisTestContainers "github.com/testcontainers/testcontainers-go/modules/redis"
)

var testConnStr string

func TestMain(m *testing.M) {
	ctx := context.Background()

	redisContainer, err := redisTestContainers.Run(ctx,
		"docker.io/redis:7.2",
		redisTestContainers.WithLogLevel(redisTestContainers.LogLevelVerbose),
	)
	if err != nil {
		panic(err)
	}

	testConnStr, err = redisContainer.ConnectionString(ctx)
	if err != nil {
		panic(err)
	}

	code := m.Run()

	_ = redisContainer.Terminate(ctx)

	os.Exit(code)
}

func setupTest(t *testing.T) (*redis.Client, *Data) {
	t.Helper()

	opts, err := redis.ParseURL(testConnStr)
	require.NoError(t, err)

	c := redis.NewClient(opts)

	d, err := NewData(testConnStr, false)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = c.FlushDB(context.Background()).Err()
		_ = c.Close()
		_ = d.Close()
	})

	return c, d
}
