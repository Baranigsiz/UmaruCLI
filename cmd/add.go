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
	Use:   "add [addons...]",
	Short: "Inject modular infrastructure addons into an existing project",
	Long: `Detects the current project type and injects modular infrastructure addons:
  - postgres : PostgreSQL connection pool & config
  - sqlite   : SQLite embedded database setup
  - jwt      : JWT authentication middleware/guard
  - redis    : Redis cache client configuration
  - docker   : Multi-stage Dockerfile, docker-compose.yml & .dockerignore
  - ci       : GitHub Actions CI/CD pipeline (.github/workflows/ci.yml)

Usage:
  umaru add                       # Interactive multi-select wizard
  umaru add redis                 # Add a single addon
  umaru add docker                # Add Docker & Compose containerization
  umaru add ci                    # Add GitHub Actions CI/CD pipeline
  umaru add postgres redis jwt    # Add multiple addons in one pass
  umaru add sqlite redis -f       # Overwrite existing addon files`,
	ValidArgs: []string{"redis", "jwt", "postgres", "sqlite", "docker", "ci"},
	Args:      cobra.ArbitraryArgs,
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

		if len(args) > 0 {
			for _, rawArg := range args {
				arg := strings.ToLower(strings.TrimSpace(rawArg))
				switch arg {
				case "ci", "github-actions", "workflow", "actions":
					addonConfig.CI = true
				case "docker", "container", "compose":
					addonConfig.Docker = true
				case "redis", "cache":
					addonConfig.Redis = true
				case "jwt", "auth":
					addonConfig.Auth = "jwt"
				case "postgres", "postgresql", "pg":
					if addonConfig.Database == "sqlite" {
						fmt.Println("❌ Cannot select both PostgreSQL and SQLite. Choose one database driver.")
						os.Exit(1)
					}
					addonConfig.Database = "postgres"
				case "sqlite", "sqlite3":
					if addonConfig.Database == "postgres" {
						fmt.Println("❌ Cannot select both PostgreSQL and SQLite. Choose one database driver.")
						os.Exit(1)
					}
					addonConfig.Database = "sqlite"
				default:
					fmt.Printf("❌ Unknown addon '%s'. Supported addons: postgres, sqlite, jwt, redis, docker, ci\n", arg)
					os.Exit(1)
				}
			}
		} else {
			// Interactive Multi-Selection
			var selectedChoices []string
			options := []huh.Option[string]{
				huh.NewOption("🐘 PostgreSQL (Connection pool & healthcheck)", "postgres"),
				huh.NewOption("📦 SQLite (Embedded file-based DB)", "sqlite"),
				huh.NewOption("🔐 JWT (Authentication middleware & claims)", "jwt"),
				huh.NewOption("🔴 Redis (In-memory caching client)", "redis"),
				huh.NewOption("🐳 Docker (Multi-stage Dockerfile & Compose)", "docker"),
				huh.NewOption("🤖 GitHub Actions CI/CD (.github/workflows/ci.yml)", "ci"),
			}

			prompt := huh.NewMultiSelect[string]().
				Title(fmt.Sprintf("Select addons to inject into %s (%s)", proj.ProjectName, proj.Framework)).
				Description("Press [Space] to toggle, [Enter] to confirm").
				Options(options...).
				Validate(func(selected []string) error {
					if len(selected) == 0 {
						return errors.New("please select at least one addon")
					}
					hasPG := false
					hasSQLite := false
					for _, s := range selected {
						if s == "postgres" {
							hasPG = true
						}
						if s == "sqlite" {
							hasSQLite = true
						}
					}
					if hasPG && hasSQLite {
						return errors.New("cannot select both PostgreSQL and SQLite. Please select only one database")
					}
					return nil
				}).
				Value(&selectedChoices)

			if err := prompt.Run(); err != nil {
				if errors.Is(err, huh.ErrUserAborted) {
					fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#94A3B8")).Render("\nOperation cancelled."))
					return
				}
				fmt.Printf("❌ %v\n", err)
				os.Exit(1)
			}

			for _, choice := range selectedChoices {
				switch choice {
				case "postgres":
					addonConfig.Database = "postgres"
				case "sqlite":
					addonConfig.Database = "sqlite"
				case "jwt":
					addonConfig.Auth = "jwt"
				case "redis":
					addonConfig.Redis = true
				case "docker":
					addonConfig.Docker = true
				case "ci":
					addonConfig.CI = true
				}
			}
		}

		if !addonConfig.HasAddons() {
			fmt.Println("No addons selected.")
			return
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
			_ = actions.InstallDependencies(generator.GetAddonBaseDir(projConfig), []string{"go", "mod", "tidy"}, false)
		}

		// Success Card
		titleStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#10B981"))

		labelStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#94A3B8"))

		valueStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F8FAFC"))

		fileStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00D8F6"))

		cmdStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#A3E635"))

		boxStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#10B981")).
			Padding(1, 2).
			MarginTop(1)

		var sb strings.Builder
		sb.WriteString(titleStyle.Render("🧩 Addon(s) Injected Successfully!") + "\n\n")

		var addonsList []string
		if addonConfig.Database != "" && addonConfig.Database != "none" {
			addonsList = append(addonsList, "DB: "+addonConfig.Database)
		}
		if addonConfig.Auth != "" && addonConfig.Auth != "none" {
			addonsList = append(addonsList, "Auth: "+addonConfig.Auth)
		}
		if addonConfig.Redis {
			addonsList = append(addonsList, "Cache: Redis")
		}
		if addonConfig.Docker {
			addonsList = append(addonsList, "Docker: Containerized")
		}
		if addonConfig.CI {
			addonsList = append(addonsList, "CI/CD: GitHub Actions")
		}

		sb.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Injected:   "), valueStyle.Render(strings.Join(addonsList, ", "))))
		sb.WriteString(fmt.Sprintf("%s %s (%s)\n", labelStyle.Render("Project:    "), valueStyle.Render(proj.ProjectName), valueStyle.Render(string(proj.Framework))))
		sb.WriteString(fmt.Sprintf("%s %s\n\n", labelStyle.Render("Target:     "), valueStyle.Render(targetDir)))

		sb.WriteString(labelStyle.Render("Generated Addon Files:") + "\n")
		for _, f := range addonFiles {
			rel, err := filepath.Rel(targetDir, f)
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
		case generator.ProjectTypeRust:
			sb.WriteString(fmt.Sprintf("  1. Run %s to verify your project dependencies and builds\n", cmdStyle.Render("cargo check")))
			sb.WriteString("  2. Review the generated configuration in your project\n")
		}

		fmt.Println(boxStyle.Render(sb.String()))
		fmt.Println()
	},
}

func init() {
	addCmd.Flags().StringVarP(&addDirFlag, "dir", "d", ".", "Target project directory")
	addCmd.Flags().BoolVarP(&addForceFlag, "force", "f", false, "Overwrite existing files if present")

	addCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		allAddons := []string{
			"postgres\tPostgreSQL connection pool & health check",
			"sqlite\tSQLite embedded lightweight database",
			"jwt\tJSON Web Token authentication middleware",
			"redis\tRedis in-memory caching client",
			"docker\tMulti-stage Dockerfile & docker-compose.yml",
			"ci\tGitHub Actions CI/CD workflow pipeline",
		}
		selected := make(map[string]bool)
		for _, arg := range args {
			selected[strings.ToLower(arg)] = true
		}
		var available []string
		for _, a := range allAddons {
			key := strings.Split(a, "\t")[0]
			if !selected[key] {
				available = append(available, a)
			}
		}
		return available, cobra.ShellCompDirectiveNoFileComp
	}

	rootCmd.AddCommand(addCmd)
}
