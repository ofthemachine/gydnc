package cmd

import (
	_ "embed"
	"fmt"

	"github.com/spf13/cobra"
)

//go:embed version.txt
var versionString string

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of gydnc",
	Long:  `All software has versions. This is gydnc's.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(versionString)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
