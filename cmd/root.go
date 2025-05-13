package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gydnc",
	Short: "A CLI for content-addressable guidance management.",
	Long:  `Gydnc is a command-line tool to create, update, and manage guidance entities using a content-addressable system.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(llmCmd) // llmCmd is defined in llm.go
}
