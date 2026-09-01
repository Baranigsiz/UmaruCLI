package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"umaru/internal/actions"
	"umaru/internal/checks"
	"umaru/internal/config"
	"umaru/internal/generator"
	"umaru/internal/prompts"
	"umaru/internal/templates"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	templateFlag       string
	packageManagerFlag string
	fromFlag           string
	dbFlag             string
	authFlag           string
	redisFlag          bool
	noAddonsFlag       bool
	noGitFlag          bool
	skipInstallFlag    bool
	forceFlag          bool
	verboseFlag        bool
	dryRunFlag         bool
)

func printBanner() {
	bannerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		MarginBottom(1)

	subStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00D8F6")).
		Bold(true)

	fmt.Println()
	fmt.Printf("%s %s\n\n", bannerStyle.Render("⚡ UMARU CLI"), subStyle.Render("Production-Ready Project Scaffolder"))
}

func printDryRunCard(config generator.ProjectConfig, templateName string, files []string) {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FBBF24"))

	labelStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#94A3B8"))

	fileStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A3E635"))

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FBBF24")).
		Padding(1, 2).
		MarginTop(1)

	var sb strings.Builder
	sb.WriteString(titleStyle.Render("🔍 Dry-Run Mode (Simulation - No files written)") + "\n\n")
	sb.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("📁 Target:   "), config.TargetDir))
	sb.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("📦 Template: "), templateName))

	if config.Addons.HasAddons() {
		var addonsList []string
		if config.Addons.Database != "" && config.Addons.Database != "none" {
			addonsList = append(addonsList, "DB: "+config.Addons.Database)
		}
		if config.Addons.Auth != "" && config.Addons.Auth != "none" {
			addonsList = append(addonsList, "Auth: "+config.Addons.Auth)
		}
		if config.Addons.Redis {
			addonsList = append(addonsList, "Cache: Redis")
		}
		sb.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("🧩 Addons:   "), strings.Join(addonsList, ", ")))
	}

	sb.WriteString(fmt.Sprintf("\n%s\n", labelStyle.Render("Files that would be generated:")))

	for _, f := range files {
		sb.WriteString(fmt.Sprintf("  📄 %s\n", fileStyle.Render(f)))
	}

	fmt.Println(boxStyle.Render(sb.String()))
	fmt.Println()
}

func printSuccessCard(config generator.ProjectConfig, templateName string, runCommand string, installCommand []string) {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#10B981"))

	labelStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#94A3B8"))

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F8FAFC"))

	cmdStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00D8F6"))

	starStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FBBF24")).
		Italic(true)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1, 2).
		MarginTop(1)

	var sb strings.Builder
	sb.WriteString(titleStyle.Render("✨ Project Scaffolding Complete!") + "\n\n")
	sb.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("📁 Project:   "), valueStyle.Render(config.SafeName)))
	sb.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("📦 Template:  "), valueStyle.Render(templateName)))
	sb.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("📍 Directory: "), valueStyle.Render(config.TargetDir)))

	if config.Addons.HasAddons() {
		var addonsList []string
		if config.Addons.Database != "" && config.Addons.Database != "none" {
			addonsList = append(addonsList, "DB: "+config.Addons.Database)
		}
		if config.Addons.Auth != "" && config.Addons.Auth != "none" {
			addonsList = append(addonsList, "Auth: "+config.Addons.Auth)
		}
		if config.Addons.Redis {
			addonsList = append(addonsList, "Cache: Redis")
		}
		sb.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("🧩 Addons:    "), valueStyle.Render(strings.Join(addonsList, ", "))))
	}

	sb.WriteString("\n" + labelStyle.Render("Next steps to get started:") + "\n")
	if config.TargetDir != "." {
		sb.WriteString(fmt.Sprintf("  1. %s\n", cmdStyle.Render(fmt.Sprintf("cd %s", config.TargetDir))))
	}

	step := 2
	if config.TargetDir == "." {
		step = 1
	}

	if skipInstallFlag && len(installCommand) > 0 {
		sb.WriteString(fmt.Sprintf("  %d. %s\n", step, cmdStyle.Render(strings.Join(installCommand, " "))))
		step++
	}
	if runCommand != "" {
		sb.WriteString(fmt.Sprintf("  %d. %s\n", step, cmdStyle.Render(runCommand)))
	}

	sb.WriteString("\n" + starStyle.Render("⭐ Love Umaru? Give us a star: https://github.com/Baranigsiz/UmaruCLI"))

	fmt.Println(boxStyle.Render(sb.String()))
	fmt.Println()
}

var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Initialize a new project",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var initialName string
		if len(args) > 0 {
			initialName = args[0]
		}

		userCfg := config.LoadUserConfig()
		if !cmd.Flags().Changed("no-git") && !userCfg.GitInit {
			noGitFlag = true
		}

		// Handle Remote Template Flow (--from)
		if fromFlag != "" {
			if initialName == "" {
				initialName = "umaru-app"
			}
			projConfig, err := generator.ResolveProjectConfig(initialName, "remote")
			if err != nil {
				fmt.Printf("\n❌ Failed to resolve project config: %v\n", err)
				os.Exit(1)
			}
			projConfig.Author = userCfg.Author
			projConfig.License = userCfg.License

			if dryRunFlag {
				files, err := generator.DryRunRemote(fromFlag, projConfig)
				if err != nil {
					fmt.Printf("\n❌ Remote dry-run failed: %v\n", err)
					os.Exit(1)
				}
				printDryRunCard(projConfig, fmt.Sprintf("Remote (%s)", fromFlag), files)
				return
			}

			if err := generator.CheckDestination(projConfig.TargetDir, forceFlag); err != nil {
				fmt.Printf("\n❌ Destination check failed: %v\n", err)
				os.Exit(1)
			}

			if err := checks.PreFlightChecks([]string{"git"}, true, false); err != nil {
				fmt.Printf("\n❌ Pre-flight check failed: %v\n", err)
				os.Exit(1)
			}

			var remoteTmpl *templates.TemplateConfig
			var setupErr error

			if verboseFlag {
				fmt.Printf("🚀 Scaffolding %s from remote repository '%s'...\n", projConfig.SafeName, fromFlag)
				remoteTmpl, setupErr = generator.GenerateFromRemote(fromFlag, projConfig)
				if setupErr != nil {
					fmt.Printf("\n❌ Failed to generate from remote: %v\n", setupErr)
					os.Exit(1)
				}
				if !noGitFlag {
					fmt.Println("📦 Initializing Git repository...")
					if err := actions.InitGit(projConfig.TargetDir); err != nil {
						fmt.Printf("⚠️ Git init warning: %v\n", err)
					}
				}
			} else {
				err = spinner.New().
					Title(fmt.Sprintf("Scaffolding %s from remote repository...", projConfig.SafeName)).
					Action(func() {
						remoteTmpl, setupErr = generator.GenerateFromRemote(fromFlag, projConfig)
						if setupErr != nil {
							return
						}
						if !noGitFlag {
							setupErr = actions.InitGit(projConfig.TargetDir)
						}
					}).
					Run()

				if err != nil || setupErr != nil {
					if setupErr != nil {
						err = setupErr
					}
					fmt.Printf("\n❌ Failed to setup remote project:\n%v\n", err)
					os.Exit(1)
				}
			}

			runCmd := ""
			var installCmd []string
			if remoteTmpl != nil {
				runCmd = remoteTmpl.RunCommand
				installCmd = remoteTmpl.InstallCommand
			}
			printSuccessCard(projConfig, fmt.Sprintf("Remote (%s)", fromFlag), runCmd, installCmd)
			return
		}

		if initialName == "" && templateFlag == "" {
			printBanner()
		}

		initialAddons := generator.AddonConfig{
			Database: dbFlag,
			Auth:     authFlag,
			Redis:    redisFlag,
		}

		result, err := prompts.Run(initialName, templateFlag, packageManagerFlag, initialAddons, noAddonsFlag)
		if err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#94A3B8")).Render("\nOperation cancelled."))
				return
			}
			fmt.Printf("❌ %v\n", err)
			return
		}

		// Resolve safe naming and target directory
		projConfig, err := generator.ResolveProjectConfig(result.ProjectName, result.Template.ID)
		if err != nil {
			fmt.Printf("\n❌ Failed to resolve project config: %v\n", err)
			os.Exit(1)
		}
		projConfig.Author = userCfg.Author
		projConfig.License = userCfg.License
		projConfig.Addons = result.Addons

		// Handle Dry-Run mode
		if dryRunFlag {
			files, err := generator.DryRun(projConfig)
			if err != nil {
				fmt.Printf("\n❌ Dry-run failed: %v\n", err)
				os.Exit(1)
			}
			printDryRunCard(projConfig, result.Template.Name, files)
			return
		}

		// Check target destination directory
		if err := generator.CheckDestination(projConfig.TargetDir, forceFlag); err != nil {
			fmt.Printf("\n❌ Destination check failed: %v\n", err)
			os.Exit(1)
		}

		installCmd := result.Template.GetInstallCommand(result.PackageManager)
		runCmd := result.Template.GetRunCommand(result.PackageManager)

		// Run pre-flight checks
		if err := checks.PreFlightChecks(installCmd, !noGitFlag, !skipInstallFlag); err != nil {
			fmt.Printf("\n❌ Pre-flight check failed: %v\n", err)
			fmt.Println("Please install the missing dependency or use flags (e.g. --no-git, --skip-install) to skip.")
			os.Exit(1)
		}

		var setupErr error

		if verboseFlag {
			fmt.Printf("🚀 Scaffolding %s using %s template...\n", projConfig.SafeName, result.Template.Name)
			if err := generator.Generate(projConfig); err != nil {
				fmt.Printf("\n❌ Failed to generate project files: %v\n", err)
				os.Exit(1)
			}
			if !noGitFlag {
				fmt.Println("📦 Initializing Git repository...")
				if err := actions.InitGit(projConfig.TargetDir); err != nil {
					fmt.Printf("⚠️ Git init warning: %v\n", err)
				}
			}
			if !skipInstallFlag && len(installCmd) > 0 {
				fmt.Printf("📥 Installing dependencies with '%s'...\n", strings.Join(installCmd, " "))
				if err := actions.InstallDependencies(projConfig.TargetDir, installCmd, true); err != nil {
					fmt.Printf("\n❌ Failed to install dependencies: %v\n", err)
					os.Exit(1)
				}
			}
		} else {
			err = spinner.New().
				Title(fmt.Sprintf("Scaffolding %s using %s template...", projConfig.SafeName, result.Template.Name)).
				Action(func() {
					// 1. Generate files (including Addons)
					setupErr = generator.Generate(projConfig)
					if setupErr != nil {
						return
					}

					// 2. Init Git (unless --no-git)
					if !noGitFlag {
						setupErr = actions.InitGit(projConfig.TargetDir)
						if setupErr != nil {
							return
						}
					}

					// 3. Install dependencies (unless --skip-install)
					if !skipInstallFlag {
						setupErr = actions.InstallDependencies(projConfig.TargetDir, installCmd, false)
					}
				}).
				Run()

			if err != nil || setupErr != nil {
				if setupErr != nil {
					err = setupErr
				}
				fmt.Printf("\n❌ Failed to setup project:\n%v\n", err)
				os.Exit(1)
			}
		}

		printSuccessCard(projConfig, result.Template.Name, runCmd, installCmd)
	},
}

func init() {
	initCmd.Flags().StringVarP(&templateFlag, "template", "t", "", "Template ID to use (e.g. go-fiber, react-vite-ts)")
	initCmd.Flags().StringVarP(&packageManagerFlag, "package-manager", "p", "", "Package manager for Node templates (npm, pnpm, yarn, bun)")
	initCmd.Flags().StringVar(&fromFlag, "from", "", "Scaffold project directly from a Git repository or GitHub shorthand (e.g. owner/repo)")
	initCmd.Flags().StringVar(&dbFlag, "db", "", "Database addon driver (postgres, sqlite, mongodb, none)")
	initCmd.Flags().StringVar(&authFlag, "auth", "", "Authentication addon (jwt, none)")
	initCmd.Flags().BoolVar(&redisFlag, "redis", false, "Include Redis caching client addon")
	initCmd.Flags().BoolVar(&noAddonsFlag, "no-addons", false, "Skip interactive addon wizard")
	initCmd.Flags().BoolVar(&noGitFlag, "no-git", false, "Skip git repository initialization")
	initCmd.Flags().BoolVar(&skipInstallFlag, "skip-install", false, "Skip installing dependencies")
	initCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Overwrite existing files in target directory")
	initCmd.Flags().BoolVarP(&verboseFlag, "verbose", "v", false, "Show detailed installation command logs")
	initCmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Simulate project generation without writing files")

	// Dynamic Shell Autocompletions for Flags
	_ = initCmd.RegisterFlagCompletionFunc("template", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		tmpls, err := templates.GetAvailableTemplates()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}
		var completions []string
		for _, t := range tmpls {
			completions = append(completions, fmt.Sprintf("%s\t%s", t.ID, t.Name))
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	})

	_ = initCmd.RegisterFlagCompletionFunc("package-manager", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{
			"npm\tStandard Node Package Manager",
			"pnpm\tFast, disk space efficient package manager",
			"yarn\tClassic Yarn package manager",
			"bun\tUltra-fast all-in-one JavaScript runtime",
		}, cobra.ShellCompDirectiveNoFileComp
	})

	_ = initCmd.RegisterFlagCompletionFunc("db", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{
			"postgres\tPostgreSQL relational database",
			"sqlite\tSQLite embedded lightweight database",
			"none\tNo database driver",
		}, cobra.ShellCompDirectiveNoFileComp
	})

	_ = initCmd.RegisterFlagCompletionFunc("auth", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{
			"jwt\tJSON Web Token authentication",
			"none\tPublic / No authentication middleware",
		}, cobra.ShellCompDirectiveNoFileComp
	})

	rootCmd.AddCommand(initCmd)
}
