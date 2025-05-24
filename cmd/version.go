package cmd

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

//go:embed version.txt
var versionString string

// Build-time version information
var (
	buildVersion   = "dev"
	buildCommit    = "unknown"
	buildTimestamp = "unknown"
)

// SetVersionInfo sets the version information from build-time ldflags
func SetVersionInfo(version, commit, buildTime string) {
	if version != "" {
		buildVersion = version
	}
	if commit != "" {
		buildCommit = commit
	}
	if buildTime != "" {
		buildTimestamp = buildTime
	}
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of gydnc",
	Long:  `All software has versions. This is gydnc's.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Use embedded version.txt as primary source, fallback to build-time info
		version := strings.TrimSpace(versionString)
		if version == "" || version == "dev-version" {
			// Fallback to build-time version info
			if buildVersion != "dev" {
				version = buildVersion
			} else {
				version = "dev-version"
			}
		}

		// If verbose flag is set, show detailed information like docker version
		if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
			fmt.Printf("Client:\n")
			fmt.Printf(" Version:    %s\n", version)
			fmt.Printf(" Git commit: %s\n", buildCommit)
			fmt.Printf(" Built:      %s\n", buildTimestamp)
			fmt.Printf(" Go version: %s\n", "go1.24.2") // Could be made dynamic if needed
			fmt.Printf(" OS/Arch:    linux/amd64\n")    // Could be made dynamic if needed
		} else {
			// Default behavior - just print version
			fmt.Println(version)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.Flags().BoolP("verbose", "v", false, "Show detailed version information")
}
