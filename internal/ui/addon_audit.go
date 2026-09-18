package ui

import (
	"fmt"
	"strings"
	"umaru/internal/generator"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// PrintAddonAuditCard renders the project addon audit table
func PrintAddonAuditCard(audit *generator.ProjectAddonAudit) {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		MarginBottom(1)

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	rowStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#E2E8F0")).
		Padding(0, 1)

	addonNameStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00D8F6"))

	installedStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#10B981"))

	missingStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#94A3B8"))

	fileStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F8FAFC"))

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return headerStyle
			}
			return rowStyle
		}).
		Headers("ADDON", "STATUS", "DETECTED FILES")

	installedCount := 0
	for _, a := range audit.Addons {
		statusStr := missingStyle.Render("— Available")
		filesStr := missingStyle.Render("(none)")
		if a.Installed {
			installedCount++
			statusStr = installedStyle.Render("✔ Installed")
			filesStr = fileStyle.Render(strings.Join(a.DetectedFiles, ", "))
		}
		t.Row(
			addonNameStyle.Render(a.Name),
			statusStr,
			filesStr,
		)
	}

	title := fmt.Sprintf("🧩 Addon Audit for %s (%s)", audit.ProjectName, audit.Framework)
	fmt.Println()
	fmt.Println(titleStyle.Render(title))
	fmt.Println(t)
	if installedCount == len(audit.Addons) {
		fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981")).Render("✨ All available infrastructure addons are installed!"))
	} else {
		fmt.Println(lipgloss.NewStyle().Faint(true).Render("Usage: umaru add <addon...> to inject available components"))
	}
	fmt.Println()
}
