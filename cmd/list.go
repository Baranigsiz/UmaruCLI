package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"umaru/internal/templates"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"
)

var (
	listCategoryFlag string
	listSearchFlag   string
	listJSONFlag     bool
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all available project templates",
	GroupID: "core",
	RunE: func(cmd *cobra.Command, args []string) error {
		allTemplates, err := templates.GetAvailableTemplates()
		if err != nil {
			return fmt.Errorf("failed to load templates: %w", err)
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

		if listSearchFlag != "" {
			query := strings.ToLower(strings.TrimSpace(listSearchFlag))
			var filtered []templates.TemplateConfig
			for _, tmpl := range allTemplates {
				if strings.Contains(strings.ToLower(tmpl.ID), query) ||
					strings.Contains(strings.ToLower(tmpl.Name), query) ||
					strings.Contains(strings.ToLower(tmpl.Description), query) ||
					strings.Contains(strings.ToLower(tmpl.GetCategory()), query) {
					filtered = append(filtered, tmpl)
				}
			}
			allTemplates = filtered
		}

		if listJSONFlag {
			data, err := json.MarshalIndent(allTemplates, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to serialize templates to json: %w", err)
			}
			fmt.Println(string(data))
			return nil
		}

		if len(allTemplates) == 0 {
			var conditions []string
			if listCategoryFlag != "" {
				conditions = append(conditions, fmt.Sprintf("category '%s'", listCategoryFlag))
			}
			if listSearchFlag != "" {
				conditions = append(conditions, fmt.Sprintf("query '%s'", listSearchFlag))
			}
			if len(conditions) > 0 {
				fmt.Printf("No templates found matching %s.\n", strings.Join(conditions, " and "))
			} else {
				fmt.Println("No templates found.")
			}
			return nil
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
		var subTitles []string
		if listCategoryFlag != "" {
			catDisplay := listCategoryFlag
			if len(catDisplay) > 0 {
				catDisplay = strings.ToUpper(catDisplay[:1]) + strings.ToLower(catDisplay[1:])
			}
			subTitles = append(subTitles, catDisplay)
		}
		if listSearchFlag != "" {
			subTitles = append(subTitles, fmt.Sprintf("query: %q", listSearchFlag))
		}
		if len(subTitles) > 0 {
			title = fmt.Sprintf("📦 Available Starter Templates (%s)", strings.Join(subTitles, ", "))
		}

		fmt.Println()
		fmt.Println(titleStyle.Render(title))
		fmt.Println(t)
		fmt.Println(lipgloss.NewStyle().Faint(true).Render("Usage: umaru init <project-name> --template <id>"))
		fmt.Println()
		return nil
	},
}

func init() {
	listCmd.Flags().StringVarP(&listCategoryFlag, "category", "c", "", "Filter templates by category (Frontend, Backend, Fullstack, CLI, Desktop)")
	listCmd.Flags().StringVarP(&listSearchFlag, "search", "s", "", "Search templates by keyword, ID, or description")
	listCmd.Flags().BoolVar(&listJSONFlag, "json", false, "Output templates list in JSON format")
	_ = listCmd.RegisterFlagCompletionFunc("category", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{
			"Frontend\tWeb client applications (React, Vue, Svelte, Next.js, Astro)",
			"Backend\tServer APIs (Go, Node.js, Python, Rust)",
			"Fullstack\tMonorepos and integrated stacks",
			"CLI\tTerminal tools and command-line utilities",
			"Desktop\tCross-platform native desktop apps (Tauri, Rust, Webview)",
		}, cobra.ShellCompDirectiveNoFileComp
	})
	rootCmd.AddCommand(listCmd)
}
