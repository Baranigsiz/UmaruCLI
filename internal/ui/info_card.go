package ui

import (
	"fmt"
	"strings"
	"umaru/internal/templates"

	"github.com/charmbracelet/lipgloss"
)

// PrintTemplateInfoCard renders the detailed template inspection card
func PrintTemplateInfoCard(info *templates.TemplateInfo) {
	// Color tokens
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	idBadgeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00D8F6")).
		Background(lipgloss.Color("#1E1E2E")).
		Padding(0, 1)

	catBadgeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FBBF24")).
		Background(lipgloss.Color("#1E1E2E")).
		Padding(0, 1)

	sectionHeaderStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00D8F6")).
		MarginTop(1)

	labelStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#94A3B8"))

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#E2E8F0"))

	cmdStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#A3E635"))

	treeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#94A3B8"))

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1, 2).
		MarginTop(1).
		MarginBottom(1)

	var sb strings.Builder

	// 1. Header Line
	sb.WriteString(fmt.Sprintf("%s  %s  %s\n\n",
		titleStyle.Render("📦 "+info.Config.Name),
		catBadgeStyle.Render(info.Config.GetCategory()),
		idBadgeStyle.Render("id: "+info.Config.ID),
	))

	// 2. Description
	sb.WriteString(labelStyle.Render("Description: "))
	sb.WriteString(valueStyle.Render(info.Config.Description))
	sb.WriteString("\n\n")

	// 3. Ports & Addons
	sb.WriteString(labelStyle.Render("🌐 Network Ports:      "))
	sb.WriteString(valueStyle.Render(strings.Join(info.Ports, ", ")))
	sb.WriteByte('\n')
	sb.WriteString(labelStyle.Render("🧩 Compatible Addons:  "))
	sb.WriteString(valueStyle.Render(strings.Join(info.SupportedAddons, ", ")))
	sb.WriteByte('\n')

	// 4. Commands
	if len(info.Config.InstallCommand) > 0 {
		sb.WriteString(labelStyle.Render("📦 Install Command:    "))
		sb.WriteString(cmdStyle.Render(strings.Join(info.Config.InstallCommand, " ")))
		sb.WriteByte('\n')
	}
	if info.Config.RunCommand != "" {
		sb.WriteString(labelStyle.Render("⚡ Run Command:        "))
		sb.WriteString(cmdStyle.Render(info.Config.RunCommand))
		sb.WriteByte('\n')
	}

	// 5. File Architecture Tree
	sb.WriteString(sectionHeaderStyle.Render(fmt.Sprintf("🌲 Architecture Tree (%d files):", info.TotalFiles)))
	sb.WriteByte('\n')
	sb.WriteString(treeStyle.Render(info.FileTree))

	// 6. Next Steps Hint
	scaffoldCmd := fmt.Sprintf("umaru init my-%s -t %s", info.Config.ID, info.Config.ID)
	sb.WriteByte('\n')
	sb.WriteString(labelStyle.Render("🚀 Quick Scaffold Command:"))
	sb.WriteByte('\n')
	sb.WriteString("  ")
	sb.WriteString(cmdStyle.Render(scaffoldCmd))
	sb.WriteByte('\n')

	fmt.Println()
	fmt.Println(boxStyle.Render(sb.String()))
}
