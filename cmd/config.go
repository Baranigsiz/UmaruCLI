package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"umaru/internal/config"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"
)

var configJSONFlag bool

var configCmd = &cobra.Command{
	Use:     "config",
	Short:   "Manage persistent global user preferences",
	GroupID: "config",
	Example: `  umaru config init
  umaru config list
  umaru config set package-manager pnpm
  umaru config set author "Baran Igsiz"
  umaru config get license
  umaru config reset`,
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
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.LoadUserConfig()
		if configJSONFlag {
			data, err := json.MarshalIndent(cfg, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to serialize config to json: %w", err)
			}
			fmt.Println(string(data))
			return nil
		}

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
		return nil
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get the value of a configuration key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.LoadUserConfig()
		key := strings.ToLower(strings.TrimSpace(args[0]))

		switch key {
		case "package-manager", "pm", "packagemanager":
			fmt.Println(cfg.PackageManager)
		case "author":
			fmt.Println(cfg.Author)
		case "license":
			fmt.Println(cfg.License)
		case "git-init", "gitinit", "git":
			fmt.Printf("%t\n", cfg.GitInit)
		default:
			return fmt.Errorf("unknown configuration key '%s'", key)
		}
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration key",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		cfg, err := config.SetConfigValue(key, value)
		if err != nil {
			return err
		}

		successStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#10B981"))
		keyStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00D8F6"))

		fmt.Println()
		fmt.Println(successStyle.Render(fmt.Sprintf("✔ Configuration '%s' updated successfully!", keyStyle.Render(key))))
		_ = cfg
		return nil
	},
}

var configUnsetCmd = &cobra.Command{
	Use:   "unset <key>",
	Short: "Unset a configuration key back to its default value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		_, err := config.UnsetConfigValue(key)
		if err != nil {
			return err
		}

		successStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#10B981"))
		keyStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00D8F6"))

		fmt.Println()
		fmt.Println(successStyle.Render(fmt.Sprintf("✔ Configuration '%s' unset successfully!", keyStyle.Render(key))))
		return nil
	},
}

var configInitCmd = &cobra.Command{
	Use:     "init",
	Short:   "Interactive setup wizard to initialize global user configuration",
	Example: `  umaru config init`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !isTerminalStdin() {
			return errors.New("interactive config setup requires a terminal standard input")
		}

		current := config.LoadUserConfig()

		author := current.Author
		pkgManager := current.PackageManager
		if pkgManager == "" {
			pkgManager = "npm"
		}
		license := current.License
		if license == "" {
			license = "MIT"
		}
		gitInit := current.GitInit

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Default Project Author").
					Description("Your name or organization handle (optional)").
					Value(&author),
				huh.NewSelect[string]().
					Title("Default Package Manager").
					Description("Preferred package manager for JavaScript & TypeScript templates").
					Options(
						huh.NewOption("npm (standard Node.js)", "npm"),
						huh.NewOption("pnpm (fast, efficient disk space)", "pnpm"),
						huh.NewOption("bun (modern, all-in-one)", "bun"),
						huh.NewOption("yarn (classic)", "yarn"),
						huh.NewOption("None (prompt interactively each time)", ""),
					).
					Value(&pkgManager),
				huh.NewSelect[string]().
					Title("Default Open Source License").
					Description("Default license added when creating new projects").
					Options(
						huh.NewOption("MIT License (Permissive, recommended)", "MIT"),
						huh.NewOption("Apache 2.0 (Includes patent rights grant)", "Apache-2.0"),
						huh.NewOption("GPL 3.0 (Strong copyleft)", "GPL-3.0"),
						huh.NewOption("BSD 3-Clause (Permissive with non-endorsement)", "BSD-3-Clause"),
						huh.NewOption("ISC License (Simplified MIT equivalent)", "ISC"),
						huh.NewOption("The Unlicense (Public domain dedication)", "Unlicense"),
					).
					Value(&license),
				huh.NewConfirm().
					Title("Automatically initialize Git repository?").
					Description("Runs 'git init' automatically during project scaffolding").
					Value(&gitInit),
			),
		)

		if err := form.Run(); err != nil {
			return err
		}

		updated := config.UserConfig{
			Author:         strings.TrimSpace(author),
			PackageManager: pkgManager,
			License:        license,
			GitInit:        gitInit,
		}

		if err := config.SaveUserConfig(updated); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		successStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#10B981"))
		titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))

		fmt.Println()
		fmt.Println(titleStyle.Render("⚙️ Umaru CLI Configuration Wizard"))
		fmt.Println(successStyle.Render("✔ Global configuration initialized successfully!"))
		fmt.Println()

		return nil
	},
}

var configResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset all configuration keys to default",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.ResetConfig(); err != nil {
			return fmt.Errorf("failed to reset configuration: %w", err)
		}

		successStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#10B981"))
		fmt.Println()
		fmt.Println(successStyle.Render("✔ Global configuration reset to default settings."))
		fmt.Println()
		return nil
	},
}

var configKeyCompletions = []string{
	"package-manager\tDefault JS/TS package manager (npm, pnpm, yarn, bun)",
	"author\tDefault project author name",
	"license\tDefault project license (e.g. MIT)",
	"git-init\tAuto initialize git repository (true/false)",
}

func init() {
	configGetCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return configKeyCompletions, cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	configUnsetCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return configKeyCompletions, cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	configSetCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return configKeyCompletions, cobra.ShellCompDirectiveNoFileComp
		}
		if len(args) == 1 {
			switch strings.ToLower(args[0]) {
			case "package-manager", "pm", "packagemanager":
				return []string{"npm", "pnpm", "yarn", "bun"}, cobra.ShellCompDirectiveNoFileComp
			case "git-init", "git", "gitinit":
				return []string{"true", "false"}, cobra.ShellCompDirectiveNoFileComp
			case "license":
				return []string{"MIT", "Apache-2.0", "GPL-3.0", "BSD-3-Clause", "ISC", "Unlicense"}, cobra.ShellCompDirectiveNoFileComp
			}
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	configListCmd.Flags().BoolVar(&configJSONFlag, "json", false, "Output configuration in JSON format")
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configUnsetCmd)
	configCmd.AddCommand(configResetCmd)
	rootCmd.AddCommand(configCmd)
}
