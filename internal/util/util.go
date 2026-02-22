// Package util provides utility functions for the application.
package util

import (
	"math/rand"
	"net/url"
	"os"
	"strconv"
	"time"
)

// Logfile is the file used for debug logging.
var Logfile *os.File

// Debug writes debug messages to the log file if debug mode is enabled.
func Debug(a ...string) {
	if Logfile == nil {
		return
	}

	PanicOnError(Logfile.WriteString(time.Now().UTC().Format("2006-01-02 15:04:05 ")))
	for _, s := range a {
		PanicOnError(Logfile.WriteString(s))
	}

	PanicOnError(Logfile.WriteString("\n"))
}

func randFloat(lower, upper float64) float64 {
	return lower + rand.Float64()*(upper-lower) //nolint:gosec // not security-sensitive, used only for debug delay
}

// DebugDelay introduces an artificial delay for testing purposes.
func DebugDelay(randomness float64) {
	if v, ok := os.LookupEnv("DEBUG_DELAY"); ok {
		if delay, err := strconv.Atoi(v); err == nil {
			if randomness > 0 {
				delay = int(randFloat(float64(delay)*(1-randomness), float64(delay)*(1+randomness)))
			}
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
	}
}

// PanicOnError returns the value if err is nil, otherwise panics.
func PanicOnError[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// NormalizeURI returns a URI that matches the format expected by go-redis's URI parsing.
//
// redis-cli expects a URI in the format "redis[s]://[password@]host[:port]", but go-redis's URI parsing expects
// a colon before the password, if password is present. The parsing in go-redis is standard parsing; the omitted colon
// expected by redis-cli is non-standard, but well-intended, since redis supports passwords but not usernames.
func NormalizeURI(uri string) (string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", err
	}
	if u.User != nil && u.User.Username() != "" {
		// if the username is set, the value is the password
		u.User = url.UserPassword("", u.User.Username())
	}
	return u.String(), nil
}
