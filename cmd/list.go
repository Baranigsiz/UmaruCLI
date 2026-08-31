package cmd

import (
	"fmt"
	"os"

	"umaru/internal/templates"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available project templates",
	Run: func(cmd *cobra.Command, args []string) {
		allTemplates, err := templates.GetAvailableTemplates()
		if err != nil {
			fmt.Printf("❌ Failed to load templates: %v\n", err)
			os.Exit(1)
		}

		if len(allTemplates) == 0 {
			fmt.Println("No templates found.")
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
			Headers("ID", "NAME", "DESCRIPTION", "RUN COMMAND")

		for _, tmpl := range allTemplates {
			t.Row(
				idStyle.Render(tmpl.ID),
				tmpl.Name,
				tmpl.Description,
				runCmdStyle.Render(tmpl.RunCommand),
			)
		}

		titleStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			MarginBottom(1)

		fmt.Println()
		fmt.Println(titleStyle.Render("📦 Available Starter Templates"))
		fmt.Println(t)
		fmt.Println(lipgloss.NewStyle().Faint(true).Render("Usage: umaru init <project-name> --template <id>"))
		fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
