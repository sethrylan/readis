package data

import (
	"context"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
	"github.com/sethrylan/readis/internal/util"
)

// Scan providers functions executing a redis Scan operation and receiving incremental results.
type Scan struct {
	pageSize int
	pattern  string
	scanning bool
	iters    map[string]*redis.ScanIterator
}

// NewScan creates a new Scan instance with the given pattern and page size.
func NewScan(pattern string, pageSize int) *Scan {
	util.Debug("new scan: ", pattern, strconv.Itoa(pageSize))
	return &Scan{
		pageSize: pageSize,
		pattern:  pattern,
		iters:    make(map[string]*redis.ScanIterator),
	}
}

// Scanning returns true if a scan is currently in progress.
func (s *Scan) Scanning() bool {
	return s.scanning
}

// HasMore returns true if there may be more keys to scan.
func (s *Scan) HasMore() bool {
	return strings.Contains(s.pattern, "*")
}

// PipelinedCmds executes pipelined commands to fetch key metadata.
func (s *Scan) PipelinedCmds(ctx context.Context, rc *redis.Client) ([]redis.Cmder, error) {
	var numFound int

	if s.iters[rc.Options().Addr] == nil {
		util.Debug("new iterator: ", rc.Options().Addr)
		s.iters[rc.Options().Addr] = rc.Scan(ctx, 0, s.pattern, int64(s.pageSize)).Iterator()
	}
	iter := s.iters[rc.Options().Addr]

	return rc.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for iter.Next(ctx) && numFound < s.pageSize {
			numFound++
			pipe.TTL(ctx, iter.Val())
			pipe.Type(ctx, iter.Val())
			pipe.MemoryUsage(ctx, iter.Val())
		}
		return iter.Err()
	})
}
