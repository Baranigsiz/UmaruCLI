package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"umaru/internal/generator"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"
)

var (
	generateDirFlag    string
	generateForceFlag  bool
	generateDryRunFlag bool
	generateJSONFlag   bool
)

var generateCmd = &cobra.Command{
	Use:     "generate [type] [name]",
	Aliases: []string{"g", "gen"},
	Short:   "Scaffold clean-architecture resources, handlers, models, and endpoints",
	GroupID: "core",
	Example: `  umaru generate resource User
  umaru g resource Product
  umaru g Order
  umaru g resource Article --dry-run
  umaru g resource Customer -f
  umaru g resource Invoice --json`,
	Long: `Inspects the current project architecture (Go, Node/TypeScript, Python, Rust)
and scaffolds end-to-end clean code boilerplate:
  - Go     : internal/models, internal/repository, internal/service, internal/handlers
  - Node   : src/models, src/services, src/controllers, src/routes
  - Python : app/schemas, app/api/endpoints (or app/routers)
  - Rust   : src/models`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var resourceType string
		var resourceName string

		switch len(args) {
		case 0:
			// Interactive mode
			if !isTerminalStdin() {
				return fmt.Errorf("interactive prompt unavailable in non-terminal mode. Specify resource name: umaru generate resource <name>")
			}

			form := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[string]().
						Title("What type of resource would you like to generate?").
						Options(
							huh.NewOption("Fullstack CRUD Resource (Model, Service, Repository, Handler)", "resource"),
						).
						Value(&resourceType),
					huh.NewInput().
						Title("Resource entity name:").
						Placeholder("e.g. User, Product, Order, Customer").
						Validate(func(s string) error {
							if strings.TrimSpace(s) == "" {
								return fmt.Errorf("name cannot be empty")
							}
							return nil
						}).
						Value(&resourceName),
				),
			)

			if err := form.Run(); err != nil {
				return err
			}
		case 1:
			// e.g. "umaru g User" -> default type is "resource"
			if strings.EqualFold(args[0], "resource") {
				if !isTerminalStdin() {
					return fmt.Errorf("please provide a resource name: umaru generate resource <name>")
				}
				inputForm := huh.NewForm(
					huh.NewGroup(
						huh.NewInput().
							Title("Resource entity name:").
							Placeholder("e.g. User, Product, Order").
							Validate(func(s string) error {
								if strings.TrimSpace(s) == "" {
									return fmt.Errorf("name cannot be empty")
								}
								return nil
							}).
							Value(&resourceName),
					),
				)
				if err := inputForm.Run(); err != nil {
					return err
				}
				resourceType = "resource"
			} else {
				resourceType = "resource"
				resourceName = args[0]
			}
		default:
			// e.g. "umaru g resource User"
			resourceType = strings.ToLower(args[0])
			resourceName = args[1]
		}

		if resourceType != "resource" {
			return fmt.Errorf("unsupported resource type '%s'. Supported types: resource", resourceType)
		}

		targetDir := "."
		if generateDirFlag != "" {
			targetDir = generateDirFlag
		}

		cfg := generator.ResourceConfig{
			Name:      resourceName,
			TargetDir: targetDir,
			Force:     generateForceFlag,
			DryRun:    generateDryRunFlag,
		}

		result, err := generator.GenerateResource(cfg)
		if err != nil {
			return err
		}

		if generateJSONFlag {
			data, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		}

		titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
		successStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#10B981"))
		warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B"))
		pathStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00D8F6"))
		tagStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#A3E635"))

		fmt.Println()
		fmt.Printf("%s  %s  %s\n\n",
			titleStyle.Render("⚡ Umaru Resource Generator:"),
			tagStyle.Render("["+result.ResourceName+"]"),
			pathStyle.Render("("+strings.ToUpper(result.Language)+" / "+result.Framework+")"),
		)

		t := table.New().
			Border(lipgloss.RoundedBorder()).
			BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))).
			Headers("GENERATED FILE", "STATUS")

		for _, f := range result.Files {
			statusRender := successStyle.Render("✔ " + f.Action)
			if strings.Contains(f.Action, "skipped") {
				statusRender = warnStyle.Render("⚠️ " + f.Action)
			}
			t.Row(f.RelPath, statusRender)
		}

		fmt.Println(t)
		fmt.Println()

		if generateDryRunFlag {
			fmt.Println(warnStyle.Render("💡 Dry-run mode: no files were written to disk. Run without --dry-run to generate."))
		} else {
			fmt.Println(successStyle.Render(fmt.Sprintf("🎉 Successfully scaffolded '%s' resource across %d files!", result.ResourceName, len(result.Files))))
		}
		fmt.Println()

		return nil
	},
}

func init() {
	generateCmd.Flags().StringVarP(&generateDirFlag, "dir", "d", ".", "Target project directory")
	generateCmd.Flags().BoolVarP(&generateForceFlag, "force", "f", false, "Overwrite existing files if they already exist")
	generateCmd.Flags().BoolVar(&generateDryRunFlag, "dry-run", false, "Preview files that would be generated without writing to disk")
	generateCmd.Flags().BoolVar(&generateJSONFlag, "json", false, "Output generation report as JSON")

	rootCmd.AddCommand(generateCmd)
}
