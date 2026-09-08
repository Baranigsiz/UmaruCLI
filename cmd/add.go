package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"umaru/internal/actions"
	"umaru/internal/generator"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	addDirFlag   string
	addForceFlag bool
)

var addCmd = &cobra.Command{
	Use:   "add [addon]",
	Short: "Inject infrastructure addons into an existing project",
	Long: `Detects the current project type and injects modular infrastructure addons:
  - postgres : PostgreSQL connection pool & config
  - sqlite   : SQLite embedded database setup
  - jwt      : JWT authentication middleware/guard
  - redis    : Redis cache client configuration

Usage:
  umaru add               # Interactive wizard
  umaru add redis         # Add Redis caching client
  umaru add jwt           # Add JWT authentication middleware
  umaru add postgres      # Add PostgreSQL database driver
  umaru add sqlite        # Add SQLite database driver`,
	ValidArgs: []string{"redis", "jwt", "postgres", "sqlite"},
	Args:      cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		targetDir := addDirFlag
		if targetDir == "" {
			targetDir = "."
		}

		proj, err := generator.DetectProject(targetDir)
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			os.Exit(1)
		}

		addonConfig := generator.AddonConfig{}
		var selectedAddonName string

		if len(args) > 0 {
			firstArg := strings.ToLower(strings.TrimSpace(args[0]))
			secondArg := ""
			if len(args) > 1 {
				secondArg = strings.ToLower(strings.TrimSpace(args[1]))
			}

			switch firstArg {
			case "redis":
				addonConfig.Redis = true
				selectedAddonName = "Redis Cache"
			case "jwt", "auth":
				addonConfig.Auth = "jwt"
				selectedAddonName = "JWT Authentication"
			case "postgres", "postgresql", "pg":
				addonConfig.Database = "postgres"
				selectedAddonName = "PostgreSQL Database"
			case "sqlite", "sqlite3":
				addonConfig.Database = "sqlite"
				selectedAddonName = "SQLite Database"
			case "db", "database":
				if secondArg == "postgres" || secondArg == "postgresql" || secondArg == "pg" {
					addonConfig.Database = "postgres"
					selectedAddonName = "PostgreSQL Database"
				} else if secondArg == "sqlite" || secondArg == "sqlite3" {
					addonConfig.Database = "sqlite"
					selectedAddonName = "SQLite Database"
				} else {
					fmt.Println("❌ Please specify a database driver: 'postgres' or 'sqlite'")
					fmt.Println("   Example: umaru add db postgres")
					os.Exit(1)
				}
			default:
				fmt.Printf("❌ Unknown addon '%s'. Supported addons: postgres, sqlite, jwt, redis\n", firstArg)
				os.Exit(1)
			}
		} else {
			// Interactive Selection
			var choiceKey string
			options := []huh.Option[string]{
				huh.NewOption("🐘 PostgreSQL (Connection pool & config)", "postgres"),
				huh.NewOption("📦 SQLite (Embedded file-based DB)", "sqlite"),
				huh.NewOption("🔐 JWT (Authentication middleware & claims)", "jwt"),
				huh.NewOption("🔴 Redis (In-memory caching client)", "redis"),
			}

			prompt := huh.NewSelect[string]().
				Title(fmt.Sprintf("Select an addon to inject into %s (%s)", proj.ProjectName, proj.Framework)).
				Options(options...).
				Value(&choiceKey)

			if err := prompt.Run(); err != nil {
				if errors.Is(err, huh.ErrUserAborted) {
					fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#94A3B8")).Render("\nOperation cancelled."))
					return
				}
				fmt.Printf("❌ %v\n", err)
				os.Exit(1)
			}

			switch choiceKey {
			case "postgres":
				addonConfig.Database = "postgres"
				selectedAddonName = "PostgreSQL Database"
			case "sqlite":
				addonConfig.Database = "sqlite"
				selectedAddonName = "SQLite Database"
			case "jwt":
				addonConfig.Auth = "jwt"
				selectedAddonName = "JWT Authentication"
			case "redis":
				addonConfig.Redis = true
				selectedAddonName = "Redis Cache"
			}
		}

		projConfig := proj.ToProjectConfig(addonConfig)
		addonFiles := generator.GetAddonFiles(projConfig)

		// Check for existing files
		if !addForceFlag {
			var existingFiles []string
			for _, f := range addonFiles {
				if _, err := os.Stat(f); err == nil {
					existingFiles = append(existingFiles, f)
				}
			}
			if len(existingFiles) > 0 {
				warnStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FBBF24"))
				fmt.Printf("\n%s\n", warnStyle.Render("⚠️ The following addon files already exist:"))
				for _, f := range existingFiles {
					fmt.Printf("  - %s\n", f)
				}
				fmt.Println("Use '--force' (-f) to overwrite existing files.")
				os.Exit(1)
			}
		}

		// Generate Addons
		if err := generator.GenerateAddons(projConfig); err != nil {
			fmt.Printf("❌ Failed to generate addon: %v\n", err)
			os.Exit(1)
		}

		// Auto-run dependency resolution if applicable
		if proj.Type == generator.ProjectTypeGo {
			_ = actions.InstallDependencies(proj.TargetDir, []string{"go", "mod", "tidy"}, false)
		}

		// Print Success Card
		printAddonSuccessCard(proj, selectedAddonName, addonFiles)
	},
}

func printAddonSuccessCard(proj *generator.DetectedProject, addonName string, files []string) {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#10B981"))
	labelStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#94A3B8"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#F8FAFC"))
	fileStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#A3E635"))
	cmdStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00D8F6"))

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1, 2).
		MarginTop(1)

	var sb strings.Builder
	sb.WriteString(titleStyle.Render("🧩 Addon Successfully Injected!") + "\n\n")
	sb.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("📁 Project:   "), valueStyle.Render(fmt.Sprintf("%s (%s)", proj.ProjectName, proj.Framework))))
	sb.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("🧩 Addon:     "), valueStyle.Render(addonName)))
	sb.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("📍 Directory: "), valueStyle.Render(proj.TargetDir)))

	sb.WriteString("\n" + labelStyle.Render("Created Files:") + "\n")
	for _, f := range files {
		rel, err := filepath.Rel(proj.TargetDir, f)
		if err != nil {
			rel = f
		}
		sb.WriteString(fmt.Sprintf("  📄 %s\n", fileStyle.Render(rel)))
	}

	sb.WriteString("\n" + labelStyle.Render("Next steps:") + "\n")
	switch proj.Type {
	case generator.ProjectTypeNode:
		sb.WriteString(fmt.Sprintf("  1. Run %s to install the updated dependencies in package.json\n", cmdStyle.Render("npm install (or pnpm/yarn/bun)")))
		sb.WriteString("  2. Review and import the generated addon module in your application\n")
	case generator.ProjectTypePython:
		sb.WriteString(fmt.Sprintf("  1. Run %s to install the updated dependencies in requirements.txt\n", cmdStyle.Render("pip install -r requirements.txt")))
		sb.WriteString("  2. Review and import the generated addon module in your application\n")
	case generator.ProjectTypeGo:
		sb.WriteString("  1. 'go mod tidy' was automatically executed\n")
		sb.WriteString("  2. Import and initialize the addon in your main entrypoint\n")
	}

	fmt.Println(boxStyle.Render(sb.String()))
	fmt.Println()
}

func init() {
	addCmd.Flags().StringVarP(&addDirFlag, "dir", "d", ".", "Target project directory")
	addCmd.Flags().BoolVarP(&addForceFlag, "force", "f", false, "Overwrite existing files if present")

	rootCmd.AddCommand(addCmd)
}
