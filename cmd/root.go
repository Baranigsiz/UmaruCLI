package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"umaru/internal/config"
	"umaru/internal/updater"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
)

var (
	noColorFlag    bool
	quietFlag      bool
	debugFlag      bool
	configFileFlag string
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

		if configFileFlag != "" {
			config.SetCustomConfigFile(configFileFlag)
		}
	},
}

// IsDebug returns whether debug mode is enabled via flag or env var
func IsDebug() bool {
	return debugFlag || os.Getenv("UMARU_DEBUG") == "1"
}

func Execute() {
	updater.CleanupOldExecutable()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	err := rootCmd.ExecuteContext(ctx)
	if err != nil {
		if IsDebug() {
			fmt.Fprintf(os.Stderr, "\n[DEBUG] Command execution failed: %+v\n", err)
		}
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = Version
	rootCmd.SetVersionTemplate(VersionTemplate())

	// Global flags
	rootCmd.PersistentFlags().BoolVar(&noColorFlag, "no-color", false, "Disable colored output (respects NO_COLOR env)")
	rootCmd.PersistentFlags().BoolVarP(&quietFlag, "quiet", "q", false, "Do not print non-essential log messages")
	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "Enable verbose debug mode with detailed error traces")
	rootCmd.PersistentFlags().StringVar(&configFileFlag, "config", "", "Location of custom configuration file (default ~/.umarurc.json)")

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
