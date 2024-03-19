package main

import (
	"math/rand"
	"net/url"
	"os"
	"strconv"
	"time"
)

var logfile *os.File

type (
	errMsg error
)

func debug(a ...string) {
	if logfile == nil {
		return
	}

	panicOnError(logfile.WriteString(time.Now().UTC().Format("2006-01-02 15:04:05 ")))
	for _, s := range a {
		panicOnError(logfile.WriteString(s))
	}

	panicOnError(logfile.WriteString("\n"))
}

func randFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func debugDelay(randomness float64) {
	if v, ok := os.LookupEnv("DEBUG_DELAY"); ok {
		if delay, err := strconv.Atoi(v); err == nil {
			if randomness > 0 {
				delay = int(randFloat(float64(delay)*(1-randomness), float64(delay)*(1+randomness)))
			}
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
	}
}

func panicOnError[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// normalizeUri returns a URI that matches the format expected by go-redis's URI parsing.
//
// redis-cli expects a URI in the format "redis[s]://[password@]host[:port]", but go-redis's URI parsing expects
// a colon before the password, if password is present. The parsing in go-redis is standard parsing; the omitted colon
// expected by redis-cli is non-standard, but well-intended, since redis supports passwords but not usernames.
func normalizeUri(uri string) string {
	u := panicOnError(url.Parse(uri))
	if u.User != nil && u.User.Username() != "" {
		// if the username is set, the value is the password
		u.User = url.UserPassword("", u.User.Username())
	}
	return u.String()
}
