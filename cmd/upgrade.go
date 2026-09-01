package cmd

import (
	"fmt"
	"os"

	"umaru/internal/updater"

	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var checkOnlyFlag bool

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Check for updates and upgrade Umaru CLI to the latest release",
	Run: func(cmd *cobra.Command, args []string) {
		titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
		successStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#10B981"))
		infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00D8F6"))
		warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FBBF24"))

		var release *updater.ReleaseInfo
		var err error

		err = spinner.New().
			Title("Checking for latest Umaru CLI releases...").
			Action(func() {
				release, err = updater.FetchLatestRelease()
			}).
			Run()

		if err != nil {
			fmt.Printf("❌ Failed to check for updates: %v\n", err)
			os.Exit(1)
		}

		current := Version
		latest := release.TagName

		fmt.Println()
		fmt.Printf("%s\n", titleStyle.Render("🔄 Umaru CLI Update Check"))
		fmt.Printf("  Current Version: %s\n", infoStyle.Render(current))
		fmt.Printf("  Latest Version:  %s\n\n", successStyle.Render(latest))

		isNewer := updater.IsNewerVersion(current, latest)

		if !isNewer {
			if current == "dev" {
				fmt.Println(warnStyle.Render("ℹ️ You are running a development build. Upgrade is not required."))
			} else {
				fmt.Println(successStyle.Render("✨ You are already using the latest version of Umaru CLI!"))
			}
			return
		}

		if checkOnlyFlag {
			fmt.Println(warnStyle.Render(fmt.Sprintf("⚡ A new version (%s) is available! Run 'umaru upgrade' to install it.", latest)))
			return
		}

		asset, err := release.FindAssetForSystem()
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			os.Exit(1)
		}

		var binaryBytes []byte
		err = spinner.New().
			Title(fmt.Sprintf("Downloading %s (%s)...", asset.Name, latest)).
			Action(func() {
				binaryBytes, err = updater.DownloadAndExtractBinary(asset.BrowserDownloadURL)
			}).
			Run()

		if err != nil {
			fmt.Printf("❌ Download failed: %v\n", err)
			os.Exit(1)
		}

		err = spinner.New().
			Title("Installing update...").
			Action(func() {
				err = updater.ReplaceCurrentExecutable(binaryBytes)
			}).
			Run()

		if err != nil {
			fmt.Printf("❌ Upgrade installation failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Println()
		fmt.Println(successStyle.Render(fmt.Sprintf("🎉 Successfully upgraded Umaru CLI to %s!", latest)))
		fmt.Println()
	},
}

func init() {
	upgradeCmd.Flags().BoolVar(&checkOnlyFlag, "check", false, "Only check if a newer version is available without installing")
	rootCmd.AddCommand(upgradeCmd)
}
