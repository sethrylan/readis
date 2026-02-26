package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sethrylan/readis/internal/data"
	"github.com/sethrylan/readis/internal/util"

	tea "charm.land/bubbletea/v2"
)

// ldflags added by goreleaser
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	exitCode := run()
	os.Exit(exitCode)
}

func run() int {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		}
	}()

	debugFlag := flag.Bool("debug", false, "Enable debug logging to the debug.log file")
	clusterFlag := flag.Bool("c", false, "Use cluster mode")
	versionFlag := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("%s (%s, built on %s)\n", version, commit, date)
		return 0
	}

	if *debugFlag {
		// all calls to fmt.Println will be written to debug.log
		util.Logfile = util.PanicOnError(tea.LogToFile("debug.log", "debug"))
		defer func() {
			_ = util.Logfile.Close()
		}()
	}

	uri := flag.Arg(0)
	if uri == "" {
		uri = "redis://localhost:6379"
	}

	d, err := data.NewData(uri, *clusterFlag)
	if err != nil {
		fmt.Printf("invalid redis URI: %s\n", err)
		return 1
	}
	p := tea.NewProgram(newModel(d))

	if _, err := p.Run(); err != nil {
		fmt.Printf("could not start program: %s\n", err)
		return 1
	}

	return 0
}
