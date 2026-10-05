package audit

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// RenderReport outputs the audit report in a polished Lipgloss terminal interface
func RenderReport(report *AuditReport, verbose bool) {
	RenderReportTo(os.Stdout, report, verbose)
}

// RenderReportTo outputs the audit report to a custom writer (useful for tests)
func RenderReportTo(w io.Writer, report *AuditReport, verbose bool) {
	if report == nil {
		return
	}

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#6366F1")).
		Padding(0, 1)

	subHeaderStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#A5B4FC"))

	metaLabelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#94A3B8"))

	metaValStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#F8FAFC"))

	passStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#10B981"))

	warnStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FBBF24"))

	failStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#EF4444"))

	recStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#CBD5E1"))

	recBulletStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#F59E0B"))

	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, headerStyle.Render(" UMARU PROJECT AUDIT "))
	_, _ = fmt.Fprintf(w, "%s %s  %s %s  %s %s\n",
		metaLabelStyle.Render("Project:"), metaValStyle.Render(report.ProjectName),
		metaLabelStyle.Render("Stack:"), metaValStyle.Render(report.Language+" / "+report.Framework),
		metaLabelStyle.Render("Path:"), metaValStyle.Render(report.TargetDir),
	)
	_, _ = fmt.Fprintln(w)

	// Build checks table
	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#475569"))).
		Headers("STATUS", "CATEGORY", "CHECK", "MESSAGE")

	for _, c := range report.Checks {
		var statusStr string
		switch c.Status {
		case StatusPass:
			statusStr = passStyle.Render("✔ PASS")
		case StatusWarn:
			statusStr = warnStyle.Render("⚠ WARN")
		case StatusFail:
			statusStr = failStyle.Render("✖ FAIL")
		default:
			statusStr = string(c.Status)
		}

		t.Row(
			statusStr,
			c.Category,
			c.Title,
			c.Message,
		)
	}

	_, _ = fmt.Fprintln(w, t.Render())

	// Health score badge
	_, _ = fmt.Fprintln(w)
	score := report.HealthScore
	var badgeStyle lipgloss.Style
	var rating string

	switch {
	case score >= 85:
		badgeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#10B981")).
			Padding(0, 2)
		rating = "EXCELLENT"
	case score >= 65:
		badgeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#1E293B")).
			Background(lipgloss.Color("#FBBF24")).
			Padding(0, 2)
		rating = "NEEDS ATTENTION"
	default:
		badgeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#EF4444")).
			Padding(0, 2)
		rating = "CRITICAL ISSUES"
	}

	_, _ = fmt.Fprintf(w, "%s  %s %s\n",
		badgeStyle.Render(fmt.Sprintf("HEALTH SCORE: %d/100", score)),
		subHeaderStyle.Render("Rating:"),
		metaValStyle.Render(rating),
	)

	// Recommendations
	if len(report.Recommendations) > 0 {
		_, _ = fmt.Fprintln(w)
		_, _ = fmt.Fprintln(w, subHeaderStyle.Render("Recommendations & Action Items:"))
		for _, rec := range report.Recommendations {
			cleanRec := strings.TrimSpace(rec)
			if cleanRec != "" {
				_, _ = fmt.Fprintf(w, "  %s %s\n", recBulletStyle.Render("•"), recStyle.Render(cleanRec))
			}
		}
	}

	_, _ = fmt.Fprintln(w)
}
