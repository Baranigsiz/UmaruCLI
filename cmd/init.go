package cmd

import (
	"fmt"
	"os"

	"umaru/internal/actions"
	"umaru/internal/checks"
	"umaru/internal/generator"
	"umaru/internal/prompts"

	"github.com/charmbracelet/huh/spinner"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new project",
	Run: func(cmd *cobra.Command, args []string) {
		result, err := prompts.Run()
		if err != nil {
			fmt.Println("Cancelled:", err)
			return
		}

		// Run pre-flight checks
		if err := checks.PreFlightChecks(result.Template); err != nil {
			fmt.Printf("\n❌ Pre-flight check failed: %v\n", err)
			fmt.Println("Please install the missing dependency and try again.")
			os.Exit(1)
		}

		fmt.Printf("\n🚀 Scaffolding %s using the %s template...\n\n", result.ProjectName, result.Template.Name)

		config := generator.ProjectConfig{
			ProjectName: result.ProjectName,
			Template:    result.Template.ID,
		}

		var setupErr error
		err = spinner.New().
			Title("Generating files & installing dependencies...").
			Action(func() {
				// 1. Generate files
				setupErr = generator.Generate(config)
				if setupErr != nil {
					return
				}

				// 2. Init Git
				setupErr = actions.InitGit(result.ProjectName)
				if setupErr != nil {
					return
				}

				// 3. Install dependencies
				setupErr = actions.InstallDependencies(result.ProjectName, result.Template)
			}).
			Run()

		if err != nil || setupErr != nil {
			if setupErr != nil {
				err = setupErr
			}
			fmt.Printf("❌ Failed to setup project: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Success! Your project is ready.\n\n")
		fmt.Printf("Next steps:\n")
		fmt.Printf("  cd %s\n", result.ProjectName)
		if result.Template.RunCommand != "" {
			fmt.Printf("  %s\n", result.Template.RunCommand)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
