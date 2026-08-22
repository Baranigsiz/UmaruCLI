package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "umaru",
	Short: "Umaru is a professional CLI tool to bootstrap developer projects",
	Long: `A fast and beautiful boilerplate generator for modern development.
Umaru helps you kickstart your projects with best practices out of the box.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Root command flags can be added here
}
