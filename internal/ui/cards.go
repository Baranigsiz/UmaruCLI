package ui

import (
	"fmt"
	"strings"
	"umaru/internal/generator"

	"github.com/charmbracelet/lipgloss"
)

// PrintBanner prints the Umaru CLI welcome banner
func PrintBanner() {
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

// PrintDryRunCard prints simulated file generation details
func PrintDryRunCard(config generator.ProjectConfig, templateName string, files []string) {
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
	sb.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("📦 Template: "), templateName))

	if config.Addons.HasAddons() {
		var addonsList []string
		if config.Addons.Database != "" && config.Addons.Database != "none" {
			addonsList = append(addonsList, "DB: "+config.Addons.Database)
		}
		if config.Addons.Auth != "" && config.Addons.Auth != "none" {
			addonsList = append(addonsList, "Auth: "+config.Addons.Auth)
		}
		if config.Addons.Redis {
			addonsList = append(addonsList, "Cache: Redis")
		}
		sb.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("🧩 Addons:   "), strings.Join(addonsList, ", ")))
	}

	sb.WriteString(fmt.Sprintf("\n%s\n", labelStyle.Render("Files that would be generated:")))

	for _, f := range files {
		sb.WriteString(fmt.Sprintf("  📄 %s\n", fileStyle.Render(f)))
	}

	fmt.Println(boxStyle.Render(sb.String()))
	fmt.Println()
}

// PrintSuccessCard renders the completion summary and next steps
func PrintSuccessCard(config generator.ProjectConfig, templateName string, runCommand string, installCommand []string, skipInstall bool) {
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
	sb.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("📍 Directory: "), valueStyle.Render(config.TargetDir)))

	if config.Addons.HasAddons() {
		var addonsList []string
		if config.Addons.Database != "" && config.Addons.Database != "none" {
			addonsList = append(addonsList, "DB: "+config.Addons.Database)
		}
		if config.Addons.Auth != "" && config.Addons.Auth != "none" {
			addonsList = append(addonsList, "Auth: "+config.Addons.Auth)
		}
		if config.Addons.Redis {
			addonsList = append(addonsList, "Cache: Redis")
		}
		sb.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("🧩 Addons:    "), valueStyle.Render(strings.Join(addonsList, ", "))))
	}

	sb.WriteString("\n" + labelStyle.Render("Next steps to get started:") + "\n")
	if config.TargetDir != "." {
		sb.WriteString(fmt.Sprintf("  1. %s\n", cmdStyle.Render(fmt.Sprintf("cd %s", config.TargetDir))))
	}

	step := 2
	if config.TargetDir == "." {
		step = 1
	}

	if skipInstall && len(installCommand) > 0 {
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
