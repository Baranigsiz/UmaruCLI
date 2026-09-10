package cmd

import (
	"fmt"
	"os"
	"strings"

	"umaru/internal/templates"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"
)

var listCategoryFlag string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available project templates",
	Run: func(cmd *cobra.Command, args []string) {
		allTemplates, err := templates.GetAvailableTemplates()
		if err != nil {
			fmt.Printf("❌ Failed to load templates: %v\n", err)
			os.Exit(1)
		}

		if listCategoryFlag != "" {
			var filtered []templates.TemplateConfig
			for _, tmpl := range allTemplates {
				if strings.EqualFold(tmpl.GetCategory(), listCategoryFlag) {
					filtered = append(filtered, tmpl)
				}
			}
			allTemplates = filtered
		}

		if len(allTemplates) == 0 {
			if listCategoryFlag != "" {
				fmt.Printf("No templates found in category '%s'. Available categories: Frontend, Backend, Fullstack, CLI\n", listCategoryFlag)
			} else {
				fmt.Println("No templates found.")
			}
			return
		}

		headerStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

		selectedRowStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E2E8F0")).
			Padding(0, 1)

		idStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00D8F6"))

		runCmdStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A3E635"))

		categoryStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FBBF24")).
			Bold(true)

		t := table.New().
			Border(lipgloss.RoundedBorder()).
			BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))).
			StyleFunc(func(row, col int) lipgloss.Style {
				switch {
				case row == 0:
					return headerStyle
				default:
					return selectedRowStyle
				}
			}).
			Headers("ID", "CATEGORY", "NAME", "RUN COMMAND")

		for _, tmpl := range allTemplates {
			t.Row(
				idStyle.Render(tmpl.ID),
				categoryStyle.Render(tmpl.GetCategory()),
				tmpl.Name,
				runCmdStyle.Render(tmpl.RunCommand),
			)
		}

		titleStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			MarginBottom(1)

		title := "📦 Available Starter Templates"
		if listCategoryFlag != "" {
			catDisplay := listCategoryFlag
			if len(catDisplay) > 0 {
				catDisplay = strings.ToUpper(catDisplay[:1]) + strings.ToLower(catDisplay[1:])
			}
			title = fmt.Sprintf("📦 Available Starter Templates (%s)", catDisplay)
		}

		fmt.Println()
		fmt.Println(titleStyle.Render(title))
		fmt.Println(t)
		fmt.Println(lipgloss.NewStyle().Faint(true).Render("Usage: umaru init <project-name> --template <id>"))
		fmt.Println()
	},
}

func init() {
	listCmd.Flags().StringVarP(&listCategoryFlag, "category", "c", "", "Filter templates by category (Frontend, Backend, Fullstack, CLI)")
	_ = listCmd.RegisterFlagCompletionFunc("category", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{
			"Frontend\tWeb client applications (React, Vue, Svelte, Next.js, Astro)",
			"Backend\tServer APIs (Go, Node.js, Python, Rust)",
			"Fullstack\tMonorepos and integrated stacks",
			"CLI\tTerminal tools and command-line utilities",
		}, cobra.ShellCompDirectiveNoFileComp
	})
	rootCmd.AddCommand(listCmd)
}
