package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"umaru/internal/actions"
	"umaru/internal/checks"
	"umaru/internal/generator"
	"umaru/internal/prompts"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	templateFlag       string
	packageManagerFlag string
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
	sb.WriteString(fmt.Sprintf("%s %s\n\n", labelStyle.Render("📦 Template: "), templateName))
	sb.WriteString(labelStyle.Render("Files that would be generated:") + "\n")

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
	sb.WriteString(fmt.Sprintf("%s %s\n\n", labelStyle.Render("📍 Directory: "), valueStyle.Render(config.TargetDir)))

	sb.WriteString(labelStyle.Render("Next steps to get started:") + "\n")
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

		if initialName == "" && templateFlag == "" {
			printBanner()
		}

		result, err := prompts.Run(initialName, templateFlag, packageManagerFlag)
		if err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#94A3B8")).Render("\nOperation cancelled."))
				return
			}
			fmt.Printf("❌ %v\n", err)
			return
		}

		// Resolve safe naming and target directory
		config, err := generator.ResolveProjectConfig(result.ProjectName, result.Template.ID)
		if err != nil {
			fmt.Printf("\n❌ Failed to resolve project config: %v\n", err)
			os.Exit(1)
		}

		// Handle Dry-Run mode
		if dryRunFlag {
			files, err := generator.DryRun(config)
			if err != nil {
				fmt.Printf("\n❌ Dry-run failed: %v\n", err)
				os.Exit(1)
			}
			printDryRunCard(config, result.Template.Name, files)
			return
		}

		// Check target destination directory
		if err := generator.CheckDestination(config.TargetDir, forceFlag); err != nil {
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
			fmt.Printf("🚀 Scaffolding %s using %s template...\n", config.SafeName, result.Template.Name)
			if err := generator.Generate(config); err != nil {
				fmt.Printf("\n❌ Failed to generate project files: %v\n", err)
				os.Exit(1)
			}
			if !noGitFlag {
				fmt.Println("📦 Initializing Git repository...")
				if err := actions.InitGit(config.TargetDir); err != nil {
					fmt.Printf("⚠️ Git init warning: %v\n", err)
				}
			}
			if !skipInstallFlag && len(installCmd) > 0 {
				fmt.Printf("📥 Installing dependencies with '%s'...\n", strings.Join(installCmd, " "))
				if err := actions.InstallDependencies(config.TargetDir, installCmd, true); err != nil {
					fmt.Printf("\n❌ Failed to install dependencies: %v\n", err)
					os.Exit(1)
				}
			}
		} else {
			err = spinner.New().
				Title(fmt.Sprintf("Scaffolding %s using %s template...", config.SafeName, result.Template.Name)).
				Action(func() {
					// 1. Generate files
					setupErr = generator.Generate(config)
					if setupErr != nil {
						return
					}

					// 2. Init Git (unless --no-git)
					if !noGitFlag {
						setupErr = actions.InitGit(config.TargetDir)
						if setupErr != nil {
							return
						}
					}

					// 3. Install dependencies (unless --skip-install)
					if !skipInstallFlag {
						setupErr = actions.InstallDependencies(config.TargetDir, installCmd, false)
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

		printSuccessCard(config, result.Template.Name, runCmd, installCmd)
	},
}

func init() {
	initCmd.Flags().StringVarP(&templateFlag, "template", "t", "", "Template ID to use (e.g. go-fiber, react-vite-ts)")
	initCmd.Flags().StringVarP(&packageManagerFlag, "package-manager", "p", "", "Package manager for Node templates (npm, pnpm, yarn, bun)")
	initCmd.Flags().BoolVar(&noGitFlag, "no-git", false, "Skip git repository initialization")
	initCmd.Flags().BoolVar(&skipInstallFlag, "skip-install", false, "Skip installing dependencies")
	initCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Overwrite existing files in target directory")
	initCmd.Flags().BoolVarP(&verboseFlag, "verbose", "v", false, "Show detailed installation command logs")
	initCmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Simulate project generation without writing files")
	rootCmd.AddCommand(initCmd)
}

