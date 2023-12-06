package main

import (
	"os"
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

	logfile.WriteString(time.Now().UTC().Format("2006-01-02 15:04:05 "))
	for _, s := range a {
		logfile.WriteString(s)
	}

	logfile.WriteString("\n")
}

func panicOnError[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
