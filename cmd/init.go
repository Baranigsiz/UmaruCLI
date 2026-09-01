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
	"umaru/internal/ui"

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

func runScaffoldWorkflow(projConfig generator.ProjectConfig, generateFn func() error, installCmd []string, noGit bool, skipInstall bool, verbose bool, templateTitle string, runCmd string) {
	if verbose {
		fmt.Printf("🚀 Scaffolding %s using %s...\n", projConfig.SafeName, templateTitle)
		if err := generateFn(); err != nil {
			fmt.Printf("\n❌ Failed to generate project files: %v\n", err)
			os.Exit(1)
		}
		if !noGit {
			fmt.Println("📦 Initializing Git repository...")
			if err := actions.InitGit(projConfig.TargetDir); err != nil {
				fmt.Printf("⚠️ Git init warning: %v\n", err)
			}
		}
		if !skipInstall && len(installCmd) > 0 {
			fmt.Printf("📥 Installing dependencies with '%s'...\n", strings.Join(installCmd, " "))
			if err := actions.InstallDependencies(projConfig.TargetDir, installCmd, true); err != nil {
				fmt.Printf("\n❌ Failed to install dependencies: %v\n", err)
				os.Exit(1)
			}
		}
	} else {
		var setupErr error
		err := spinner.New().
			Title(fmt.Sprintf("Scaffolding %s using %s...", projConfig.SafeName, templateTitle)).
			Action(func() {
				setupErr = generateFn()
				if setupErr != nil {
					return
				}

				if !noGit {
					setupErr = actions.InitGit(projConfig.TargetDir)
					if setupErr != nil {
						return
					}
				}

				if !skipInstall && len(installCmd) > 0 {
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

	ui.PrintSuccessCard(projConfig, templateTitle, runCmd, installCmd, skipInstall)
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
				ui.PrintDryRunCard(projConfig, fmt.Sprintf("Remote (%s)", fromFlag), files)
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
			generateRemote := func() error {
				tmpl, err := generator.GenerateFromRemote(fromFlag, projConfig)
				remoteTmpl = tmpl
				return err
			}

			var installCmd []string
			runCmd := ""
			if remoteTmpl != nil {
				runCmd = remoteTmpl.RunCommand
				installCmd = remoteTmpl.InstallCommand
			}

			runScaffoldWorkflow(projConfig, generateRemote, installCmd, noGitFlag, skipInstallFlag, verboseFlag, fmt.Sprintf("Remote (%s)", fromFlag), runCmd)
			return
		}

		if initialName == "" && templateFlag == "" {
			ui.PrintBanner()
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
			ui.PrintDryRunCard(projConfig, result.Template.Name, files)
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

		generateLocal := func() error {
			return generator.Generate(projConfig)
		}

		runScaffoldWorkflow(projConfig, generateLocal, installCmd, noGitFlag, skipInstallFlag, verboseFlag, result.Template.Name, runCmd)
	},
}

func init() {
	initCmd.Flags().StringVarP(&templateFlag, "template", "t", "", "Template ID to use (e.g. go-fiber, react-vite-ts)")
	initCmd.Flags().StringVarP(&packageManagerFlag, "package-manager", "p", "", "Package manager for Node templates (npm, pnpm, yarn, bun)")
	initCmd.Flags().StringVar(&fromFlag, "from", "", "Scaffold project directly from a Git repository or GitHub shorthand (e.g. owner/repo)")
	initCmd.Flags().StringVar(&dbFlag, "db", "", "Database addon driver (postgres, sqlite, mongodb, none)")
	initCmd.Flags().StringVar(&authFlag, "auth", "", "Authentication addon (jwt, none)")
	initCmd.Flags().BoolVar(&redisFlag, "redis", false, "Include Redis caching client addon")
	initCmd.Flags().BoolVar(&noAddonsFlag, "no-addons", false, "Skip interactive addon configuration wizard")
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
