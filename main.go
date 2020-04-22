package main

import (
	"fmt"
	"os"

	"github.com/moshloop/platform-cli/cmd"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if len(commit) > 8 {
		version = fmt.Sprintf("%v, commit %v, built at %v", version, commit[0:8], date)
	}
	root := cmd.GetRootCmd(version)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
