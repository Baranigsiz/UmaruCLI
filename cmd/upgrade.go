package cmd

import (
	"fmt"
	"strings"

	"umaru/internal/updater"

	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	checkOnlyFlag    bool
	forceUpgradeFlag bool
)

var upgradeCmd = &cobra.Command{
	Use:     "upgrade",
	Short:   "Check for updates and upgrade Umaru CLI to the latest release",
	GroupID: "util",
	RunE: func(cmd *cobra.Command, args []string) error {
		titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
		successStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#10B981"))
		infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00D8F6"))
		warnStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FBBF24"))

		var release *updater.ReleaseInfo
		var fetchErr error

		spinnerErr := spinner.New().
			Title("Checking for latest Umaru CLI releases...").
			Action(func() {
				release, fetchErr = updater.FetchLatestReleaseContext(cmd.Context())
			}).
			Run()

		if spinnerErr != nil {
			return fmt.Errorf("failed to check for updates: %w", spinnerErr)
		}
		if fetchErr != nil {
			return fmt.Errorf("failed to check for updates: %w", fetchErr)
		}

		current := Version
		latest := release.TagName

		fmt.Println()
		fmt.Printf("%s\n", titleStyle.Render("🔄 Umaru CLI Update Check"))
		fmt.Printf("  Current Version: %s\n", infoStyle.Render(current))
		fmt.Printf("  Latest Version:  %s\n\n", successStyle.Render(latest))

		isNewer := updater.IsNewerVersion(current, latest)

		if !isNewer && !forceUpgradeFlag {
			if current == "dev" {
				fmt.Println(warnStyle.Render("ℹ️ You are running a development build. Use 'umaru upgrade --force' to install the official release."))
			} else {
				fmt.Println(successStyle.Render("✨ You are already using the latest version of Umaru CLI! (Use --force to reinstall)"))
			}
			return nil
		}

		if checkOnlyFlag {
			fmt.Println(warnStyle.Render(fmt.Sprintf("⚡ A new version (%s) is available! Run 'umaru upgrade' to install it.", latest)))
			return nil
		}

		asset, err := release.FindAssetForSystem()
		if err != nil {
			return err
		}

		var binaryBytes []byte
		var downloadErr error

		spinnerErr = spinner.New().
			Title(fmt.Sprintf("Downloading %s (%s)...", asset.Name, latest)).
			Action(func() {
				binaryBytes, downloadErr = updater.DownloadAndExtractBinaryContext(cmd.Context(), asset.BrowserDownloadURL)
			}).
			Run()

		if spinnerErr != nil {
			return fmt.Errorf("download failed: %w", spinnerErr)
		}
		if downloadErr != nil {
			return fmt.Errorf("download failed: %w", downloadErr)
		}

		var installErr error

		spinnerErr = spinner.New().
			Title("Installing update...").
			Action(func() {
				installErr = updater.ReplaceCurrentExecutable(binaryBytes)
			}).
			Run()

		if spinnerErr != nil {
			return fmt.Errorf("upgrade installation failed: %w", spinnerErr)
		}
		if installErr != nil {
			errLower := strings.ToLower(installErr.Error())
			if strings.Contains(errLower, "permission") || strings.Contains(errLower, "access is denied") {
				fmt.Println("💡 Tip: Try running the command with administrator or sudo privileges.")
			}
			return fmt.Errorf("upgrade installation failed: %w", installErr)
		}

		fmt.Println()
		fmt.Println(successStyle.Render(fmt.Sprintf("🎉 Successfully upgraded Umaru CLI to %s!", latest)))
		fmt.Println()
		return nil
	},
}

func init() {
	upgradeCmd.Flags().BoolVar(&checkOnlyFlag, "check", false, "Only check if a newer version is available without installing")
	upgradeCmd.Flags().BoolVarP(&forceUpgradeFlag, "force", "f", false, "Force upgrade even if currently on dev or latest version")
	rootCmd.AddCommand(upgradeCmd)
}
