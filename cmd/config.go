package cmd

import (
	"fmt"
	"os"

	"umaru/internal/config"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage persistent global user preferences",
	Long: `View and modify persistent user configuration stored in ~/.umarurc.json.

Available Keys:
  package-manager (or pm)  Default package manager (npm, pnpm, yarn, bun)
  author                   Default author name in generated projects
  license                  Default license (e.g. MIT)
  git-init                 Auto-initialize git repository (true/false)`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all user configuration settings",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.LoadUserConfig()
		cfgPath, _ := config.GetConfigFilePath()

		titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
		keyStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00D8F6"))
		valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#A3E635"))
		pathStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#94A3B8")).Italic(true)

		t := table.New().
			Border(lipgloss.RoundedBorder()).
			BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))).
			Headers("KEY", "CURRENT VALUE", "DESCRIPTION")

		pmVal := cfg.PackageManager
		if pmVal == "" {
			pmVal = "(none - prompts interactively)"
		}
		authorVal := cfg.Author
		if authorVal == "" {
			authorVal = "(none)"
		}

		t.Row(keyStyle.Render("package-manager"), valStyle.Render(pmVal), "Default JS/TS package manager")
		t.Row(keyStyle.Render("author"), valStyle.Render(authorVal), "Default project author")
		t.Row(keyStyle.Render("license"), valStyle.Render(cfg.License), "Default project license")
		t.Row(keyStyle.Render("git-init"), valStyle.Render(fmt.Sprintf("%t", cfg.GitInit)), "Auto initialize git on scaffold")

		fmt.Println()
		fmt.Printf("%s\n", titleStyle.Render("⚙️ Umaru CLI Global Configuration"))
		fmt.Printf("Config file: %s\n\n", pathStyle.Render(cfgPath))
		fmt.Println(t)
		fmt.Println()
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get the value of a configuration key",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.LoadUserConfig()
		key := args[0]

		switch key {
		case "package-manager", "pm":
			fmt.Println(cfg.PackageManager)
		case "author":
			fmt.Println(cfg.Author)
		case "license":
			fmt.Println(cfg.License)
		case "git-init", "git":
			fmt.Println(cfg.GitInit)
		default:
			fmt.Printf("❌ Unknown configuration key '%s'\n", key)
			os.Exit(1)
		}
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration key",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := args[1]

		cfg, err := config.SetConfigValue(key, value)
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			os.Exit(1)
		}

		successStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#10B981"))
		keyStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00D8F6"))

		fmt.Println()
		fmt.Println(successStyle.Render(fmt.Sprintf("✔ Configuration '%s' updated successfully!", keyStyle.Render(key))))
		_ = cfg
	},
}

var configResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset all configuration keys to default",
	Run: func(cmd *cobra.Command, args []string) {
		if err := config.ResetConfig(); err != nil {
			fmt.Printf("❌ Failed to reset configuration: %v\n", err)
			os.Exit(1)
		}

		successStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#10B981"))
		fmt.Println()
		fmt.Println(successStyle.Render("✔ Global configuration reset to default settings."))
		fmt.Println()
	},
}

func init() {
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configResetCmd)
	rootCmd.AddCommand(configCmd)
}
