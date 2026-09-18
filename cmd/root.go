package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"umaru/internal/updater"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
)

var (
	noColorFlag bool
	quietFlag   bool
)

var rootCmd = &cobra.Command{
	Use:           "umaru",
	Short:         "Umaru is a professional CLI tool to bootstrap developer projects",
	Long: `A fast and beautiful boilerplate generator for modern development.
Umaru helps you kickstart your projects with best practices out of the box.`,
	SilenceUsage:  true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Respect NO_COLOR env (https://no-color.org/) or --no-color flag
		if noColorFlag || os.Getenv("NO_COLOR") != "" {
			lipgloss.SetHasDarkBackground(false)
			os.Setenv("NO_COLOR", "1")
			lipgloss.SetColorProfile(termenv.Ascii)
		}
	},
}

func Execute() {
	updater.CleanupOldExecutable()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	err := rootCmd.ExecuteContext(ctx)
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = Version
	rootCmd.SetVersionTemplate(VersionTemplate())

	// Global flags
	rootCmd.PersistentFlags().BoolVar(&noColorFlag, "no-color", false, "Disable colored output (respects NO_COLOR env)")
	rootCmd.PersistentFlags().BoolVarP(&quietFlag, "quiet", "q", false, "Do not print non-essential log messages")

	// Command groups for organized --help output
	rootCmd.AddGroup(
		&cobra.Group{ID: "core", Title: "Core Commands:"},
		&cobra.Group{ID: "config", Title: "Configuration:"},
		&cobra.Group{ID: "util", Title: "Utilities:"},
	)
}

// VersionTemplate returns the formatted version output string
func VersionTemplate() string {
	return "Umaru CLI " + Version + " (commit: " + Commit + ", built at: " + BuildDate + ")\n"
}
