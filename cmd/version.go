package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// These variables are injected at build time via -ldflags by GoReleaser
var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Umaru",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Umaru CLI %s (commit: %s, built at: %s)\n", Version, Commit, BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
