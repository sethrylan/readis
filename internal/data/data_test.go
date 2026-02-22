package data //nolint:testpackage // white-box testing of internal package

import (
	"strconv"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewData(t *testing.T) {
	c, d := setupTest(t)
	ctx := t.Context()

	for i := range 1000 {
		_, err := c.Set(ctx, "testkey:"+strconv.Itoa(i), "testvalue", 0).Result()
		require.NoError(t, err)
	}

	assert.Equal(t, int64(1000), d.TotalKeys(ctx))
}

func TestNewDataErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		uri     string
		cluster bool
	}{
		{name: "empty URI", uri: "", cluster: false},
		{name: "invalid scheme", uri: "http://localhost:6379", cluster: false},
		{name: "invalid cluster URI", uri: "http://localhost:6379", cluster: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := NewData(tt.uri, tt.cluster)
			assert.Error(t, err)
		})
	}
}

func TestFetch(t *testing.T) {
	c, d := setupTest(t)
	ctx := t.Context()

	t.Run("string", func(t *testing.T) {
		require.NoError(t, c.Set(ctx, "str-key", "hello", 0).Err())

		result, err := d.Fetch(ctx, Key{Name: "str-key", Datatype: "string"})
		require.NoError(t, err)
		assert.Equal(t, "```hello```", result)
	})

	t.Run("list", func(t *testing.T) {
		require.NoError(t, c.RPush(ctx, "list-key", "a", "b", "c").Err())

		result, err := d.Fetch(ctx, Key{Name: "list-key", Datatype: "list"})
		require.NoError(t, err)
		assert.Equal(t, "- `a`\n- `b`\n- `c`\n", result)
	})

	t.Run("set", func(t *testing.T) {
		require.NoError(t, c.SAdd(ctx, "set-key", "x", "y").Err())

		result, err := d.Fetch(ctx, Key{Name: "set-key", Datatype: "set"})
		require.NoError(t, err)
		assert.Contains(t, result, "- `x`\n")
		assert.Contains(t, result, "- `y`\n")
	})

	t.Run("zset", func(t *testing.T) {
		require.NoError(t, c.ZAdd(ctx, "zset-key",
			redis.Z{Score: 1.0, Member: "a"},
			redis.Z{Score: 2.5, Member: "b"},
		).Err())

		result, err := d.Fetch(ctx, Key{Name: "zset-key", Datatype: "zset"})
		require.NoError(t, err)
		assert.Contains(t, result, "| score | value |")
		assert.Contains(t, result, "| 1.000000 | `a` |")
		assert.Contains(t, result, "| 2.500000 | `b` |")
	})

	t.Run("hash", func(t *testing.T) {
		require.NoError(t, c.HSet(ctx, "hash-key", "beta", "2", "alpha", "1").Err())

		result, err := d.Fetch(ctx, Key{Name: "hash-key", Datatype: "hash"})
		require.NoError(t, err)
		expected := "| field | value |\n| --- | --- |\n| alpha | 1 |\n| beta | 2 |\n"
		assert.Equal(t, expected, result)
	})

	t.Run("unknown type", func(t *testing.T) {
		result, err := d.Fetch(ctx, Key{Name: "any-key", Datatype: "unknown"})
		require.NoError(t, err)
		assert.Equal(t, "Unknown data type: unknown", result)
	})
}
