package data

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanAsync(t *testing.T) {
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
	total := 1000
	for i := range total {
		_, err := c.Set(ctx, "testkey:"+strconv.Itoa(i), "testvalue", 0).Result()
		require.NoError(t, err)
	}

	assert.Equal(t, int64(total), d.TotalKeys(ctx))

	// scanning is non-deterministic, and we re-use pageSize as part of the scan count, which is really just a hint
	// to the server. For a given pageSize, we expect to get roughly that many keys per call of ScanAsync, but may get
	// slightly more or slightly less.
	tests := []struct {
		pageSize  int
		scanLoops int
	}{
		{
			pageSize:  100,
			scanLoops: 10,
		},
		{
			pageSize:  200,
			scanLoops: 5,
		},
		{
			pageSize:  1000,
			scanLoops: 1,
		},
	}

	for _, test := range tests {
		testName := fmt.Sprintf("%d loops for %d keys in pages of %d", test.scanLoops, total, test.pageSize)
		t.Run(testName, func(t *testing.T) {
			// t.Parallel()  // cannot run in parallel

			var keys []*Key
			s := NewScan("testkey:*", test.pageSize)
			assert.False(t, s.scanning)

			for i := 0; i < test.scanLoops; i++ {
				fmt.Println("starting loop", i, "keys", len(keys))

				ch := d.ScanAsync(ctx, s) // start the scan
				//time.Sleep(10 * time.Millisecond) // wait a moment for the scan to start
				assert.True(t, s.scanning)
				for key := range ch {
					keys = append(keys, key)
				}

				if i != test.scanLoops-1 {
					assert.GreaterOrEqual(t, len(keys), (i+1)*test.pageSize)
				}
			}

			// the iterator will sometimes return `Next()==false` even when there are 7 or 8 keys that should be found on the last loop
			assert.GreaterOrEqual(t, len(keys), total-9)
			assert.False(t, s.scanning)
		})
	}

	err := d.Close()
	require.NoError(t, err)
}
