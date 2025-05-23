package main

import (
	"gydnc/cmd"
)

// Version information set at build time via ldflags
var (
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"
)

func main() {
	// Set version information for the cmd package
	cmd.SetVersionInfo(version, commit, buildTime)
	cmd.Execute()
}
