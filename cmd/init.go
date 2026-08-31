package cmd

import (
	"fmt"
	"os"
	"strings"

	"umaru/internal/actions"
	"umaru/internal/checks"
	"umaru/internal/generator"
	"umaru/internal/prompts"

	"github.com/charmbracelet/huh/spinner"
	"github.com/spf13/cobra"
)

var (
	templateFlag    string
	noGitFlag       bool
	skipInstallFlag bool
	forceFlag       bool
)

var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Initialize a new project",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var initialName string
		if len(args) > 0 {
			initialName = args[0]
		}

		result, err := prompts.Run(initialName, templateFlag)
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}

		// Check target destination directory
		if err := generator.CheckDestination(result.ProjectName, forceFlag); err != nil {
			fmt.Printf("\n❌ Destination check failed: %v\n", err)
			os.Exit(1)
		}

		// Run pre-flight checks (only check tools if we're actually going to execute them)
		if err := checks.PreFlightChecks(result.Template, !noGitFlag, !skipInstallFlag); err != nil {
			fmt.Printf("\n❌ Pre-flight check failed: %v\n", err)
			fmt.Println("Please install the missing dependency or use flags (e.g. --no-git, --skip-install) to skip.")
			os.Exit(1)
		}

		fmt.Printf("\n🚀 Scaffolding %s using the %s template...\n\n", result.ProjectName, result.Template.Name)

		config := generator.ProjectConfig{
			ProjectName: result.ProjectName,
			Template:    result.Template.ID,
		}

		var setupErr error
		err = spinner.New().
			Title("Generating files & setting up project...").
			Action(func() {
				// 1. Generate files
				setupErr = generator.Generate(config)
				if setupErr != nil {
					return
				}

				// 2. Init Git (unless --no-git)
				if !noGitFlag {
					setupErr = actions.InitGit(result.ProjectName)
					if setupErr != nil {
						return
					}
				}

				// 3. Install dependencies (unless --skip-install)
				if !skipInstallFlag {
					setupErr = actions.InstallDependencies(result.ProjectName, result.Template)
				}
			}).
			Run()

		if err != nil || setupErr != nil {
			if setupErr != nil {
				err = setupErr
			}
			fmt.Printf("❌ Failed to setup project:\n%v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Success! Your project is ready.\n\n")
		fmt.Printf("Next steps:\n")
		fmt.Printf("  cd %s\n", result.ProjectName)
		if skipInstallFlag && len(result.Template.InstallCommand) > 0 {
			fmt.Printf("  %s\n", strings.Join(result.Template.InstallCommand, " "))
		}
		if result.Template.RunCommand != "" {
			fmt.Printf("  %s\n", result.Template.RunCommand)
		}
	},
}

func init() {
	initCmd.Flags().StringVarP(&templateFlag, "template", "t", "", "Template ID to use (e.g. go-fiber, react-vite-ts)")
	initCmd.Flags().BoolVar(&noGitFlag, "no-git", false, "Skip git repository initialization")
	initCmd.Flags().BoolVar(&skipInstallFlag, "skip-install", false, "Skip installing dependencies")
	initCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Overwrite existing files in target directory")
	rootCmd.AddCommand(initCmd)
}
