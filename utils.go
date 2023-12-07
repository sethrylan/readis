package main

import (
	"math/rand"
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
