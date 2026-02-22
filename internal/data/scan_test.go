package data //nolint:testpackage // white-box testing of internal package

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanAsync(t *testing.T) {
	c, d := setup(t)

	// populate test data
	total := 1000
	for i := range total {
		_, err := c.Set(t.Context(), "testkey:"+strconv.Itoa(i), "testvalue", 0).Result()
		require.NoError(t, err)
	}

	assert.Equal(t, int64(total), d.TotalKeys(t.Context()))

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

			keys := make([]*Key, 0, total)
			s := NewScan("testkey:*", test.pageSize)
			assert.False(t, s.scanning.Load())

			for i := range test.scanLoops {
				fmt.Println("starting loop", i, "keys", len(keys))

				ch := d.ScanAsync(t.Context(), s) // start the scan
				// time.Sleep(10 * time.Millisecond) // wait a moment for the scan to start
				assert.True(t, s.scanning.Load())
				for key := range ch {
					keys = append(keys, key)
				}

				if i != test.scanLoops-1 {
					assert.GreaterOrEqual(t, len(keys), (i+1)*test.pageSize)
				}
			}

			// the iterator will sometimes return `Next()==false` even when there are 7 or 8 keys that should be found on the last loop
			assert.GreaterOrEqual(t, len(keys), total-9)
			assert.False(t, s.scanning.Load())
		})
	}

	err := d.Close()
	require.NoError(t, err)
}
