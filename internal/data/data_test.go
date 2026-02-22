package data //nolint:testpackage // white-box testing of internal package

import (
	"strconv"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	redisTestContainers "github.com/testcontainers/testcontainers-go/modules/redis"
)

func TestNewData(t *testing.T) {
	c, d := setup(t)

	// populate test data
	for i := range 1000 {
		_, err := c.Set(t.Context(), "testkey:"+strconv.Itoa(i), "testvalue", 0).Result()
		require.NoError(t, err)
	}

	assert.Equal(t, int64(1000), d.TotalKeys(t.Context()))
	err := d.Close()
	require.NoError(t, err)
}

func setup(t *testing.T) (*redis.Client, *Data) {
	t.Helper()
	redisContainer, err := redisTestContainers.Run(t.Context(),
		"docker.io/redis:7.2",
		redisTestContainers.WithLogLevel(redisTestContainers.LogLevelVerbose),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, redisContainer.Terminate(t.Context()))
	})

	connStr, err := redisContainer.ConnectionString(t.Context())
	require.NoError(t, err)
	opts, err := redis.ParseURL(connStr)
	require.NoError(t, err)

	c := redis.NewClient(opts)
	t.Cleanup(func() {
		require.NoError(t, c.Close())
	})

	d, err := NewData(connStr, false)
	require.NoError(t, err)

	return c, d
}
