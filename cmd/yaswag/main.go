package main

import (
	"fmt"
	"os"

	"github.com/fathurrohman26/yaswag/internal/cli"
)

var (
	version = "unknown" // will be set during build time
	commit  = "unknown" // will be set during build time
	date    = "unknown" // will be set during build time
)

func main() {
	app := cli.New(cli.WithVersionInfo(version, commit, date))

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %+v\n", err)
		os.Exit(1)
	}
}
