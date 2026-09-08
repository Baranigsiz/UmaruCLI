package doctor

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// RenderReport outputs the diagnostic report in a beautiful Lipgloss UI
func RenderReport(report DoctorReport, verbose bool) {
	// Styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	subtitleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00D8F6"))

	sysInfoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#94A3B8"))

	sysValStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#F8FAFC"))

	okStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#10B981"))

	warnStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FBBF24"))

	missReqStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#EF4444"))

	missOptStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#64748B"))

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	cellStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#E2E8F0")).
		Padding(0, 1)

	toolNameStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#F8FAFC"))

	verStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00D8F6"))

	catStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A78BFA"))

	dimStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#64748B"))

	boxBorderColor := lipgloss.Color("#7D56F4")
	if report.TotalScore == 100 {
		boxBorderColor = lipgloss.Color("#10B981")
	}

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(boxBorderColor).
		Padding(1, 2).
		MarginTop(1)

	// 1. Header Banner
	fmt.Println()
	fmt.Printf("%s %s\n\n", titleStyle.Render("🩺 UMARU DOCTOR"), subtitleStyle.Render("System & Environment Diagnostics"))

	// 2. System Info Line
	fmt.Printf("  %s %s  •  %s %s  •  %s %s  •  %s %s\n\n",
		sysInfoStyle.Render("OS:"), sysValStyle.Render(report.System.OS),
		sysInfoStyle.Render("Arch:"), sysValStyle.Render(report.System.Arch),
		sysInfoStyle.Render("CPUs:"), sysValStyle.Render(fmt.Sprintf("%d", report.System.NumCPU)),
		sysInfoStyle.Render("Umaru:"), sysValStyle.Render(report.System.UmaruVer),
	)

	// 3. Tools Table
	headers := []string{"STATUS", "TOOL", "VERSION", "CATEGORY", "DETAILS"}
	if verbose {
		headers = append(headers, "BINARY PATH")
	}

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#475569"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return headerStyle
			}
			return cellStyle
		}).
		Headers(headers...)

	for _, tool := range report.Tools {
		var statusIcon string
		switch tool.Status {
		case StatusOk:
			statusIcon = okStyle.Render("✔ OK")
		case StatusWarning:
			statusIcon = warnStyle.Render("⚠️ WARN")
		case StatusMissing:
			if tool.Required {
				statusIcon = missReqStyle.Render("✖ MISSING")
			} else {
				statusIcon = missOptStyle.Render("– OPTIONAL")
			}
		}

		verText := "-"
		if tool.Version != "" {
			verText = verStyle.Render(tool.Version)
		}

		details := tool.Notes
		if details == "" {
			details = tool.Description
		}
		if tool.Status == StatusMissing && tool.Required {
			details = missReqStyle.Render(details)
		} else if tool.Status == StatusWarning {
			details = warnStyle.Render(details)
		} else {
			details = dimStyle.Render(details)
		}

		row := []string{
			statusIcon,
			toolNameStyle.Render(tool.Name),
			verText,
			catStyle.Render(tool.Category),
			details,
		}

		if verbose {
			p := tool.Path
			if p == "" {
				p = "-"
			}
			row = append(row, dimStyle.Render(p))
		}

		t.Row(row...)
	}

	fmt.Println(t)
	fmt.Println()

	// 4. Template Readiness Matrix Card
	var sb strings.Builder
	matrixTitle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F8FAFC"))
	scoreColor := lipgloss.Color("#10B981")
	if report.TotalScore < 60 {
		scoreColor = lipgloss.Color("#EF4444")
	} else if report.TotalScore < 90 {
		scoreColor = lipgloss.Color("#FBBF24")
	}
	scoreStyle := lipgloss.NewStyle().Bold(true).Foreground(scoreColor)

	sb.WriteString(matrixTitle.Render("📦 Template Ecosystem Readiness: ") + scoreStyle.Render(fmt.Sprintf("%d%%", report.TotalScore)) + "\n\n")

	for _, tmpl := range report.Templates {
		var icon string
		if tmpl.IsReady {
			icon = okStyle.Render("✔")
		} else if tmpl.Ready > 0 {
			icon = warnStyle.Render("⚠️")
		} else {
			icon = missOptStyle.Render("✖")
		}

		statusDesc := fmt.Sprintf("%d/%d ready", tmpl.Ready, tmpl.Total)
		if len(tmpl.Missing) > 0 {
			statusDesc += fmt.Sprintf(" (missing: %s)", strings.Join(tmpl.Missing, ", "))
		}

		sb.WriteString(fmt.Sprintf("  %s %-32s %s\n",
			icon,
			lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#E2E8F0")).Render(tmpl.Category),
			dimStyle.Render(statusDesc),
		))
	}

	// 5. Actionable Tips / Recommendations
	var tips []string
	for _, tool := range report.Tools {
		if tool.Status == StatusMissing && tool.Required {
			tips = append(tips, fmt.Sprintf("%s is required: install from %s", tool.Name, tool.InstallTip))
		} else if tool.Status == StatusMissing && tool.InstallTip != "" {
			// Suggest for missing useful runtimes
			if tool.Name == "Cargo (Rust)" || tool.Name == "Python" || tool.Name == "pnpm" || tool.Name == "Docker CLI" {
				tips = append(tips, fmt.Sprintf("Install %s (%s) to enable related templates", tool.Name, tool.InstallTip))
			}
		} else if tool.Status == StatusWarning {
			tips = append(tips, fmt.Sprintf("%s: %s (%s)", tool.Name, tool.Notes, tool.InstallTip))
		}
	}

	if len(tips) > 0 {
		sb.WriteString("\n" + lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FBBF24")).Render("💡 Recommendations:") + "\n")
		for _, tip := range tips {
			sb.WriteString(fmt.Sprintf("  • %s\n", tip))
		}
	}

	fmt.Println(cardStyle.Render(sb.String()))

	// Final summary verdict
	fmt.Println()
	if report.TotalScore == 100 {
		fmt.Println(okStyle.Render("  🎉 Phenomenal! Your system is 100% equipped to run every Umaru starter template and addon."))
	} else if report.TotalScore >= 70 {
		fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#00D8F6")).Render("  ✨ Great setup! Most Umaru templates are ready to scaffold right out of the box."))
	} else {
		fmt.Println(warnStyle.Render("  ⚠️  Some key developer runtimes are missing. Check the recommendations above to unlock more templates."))
	}
	fmt.Println()
}
