package deploy

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// RenderResult renders the deployment results in a polished terminal UI
func RenderResult(result *DeployResult, dryRun bool) {
	RenderResultTo(os.Stdout, result, dryRun)
}

// RenderResultTo renders deployment results to an io.Writer (testable)
func RenderResultTo(w io.Writer, result *DeployResult, dryRun bool) {
	if result == nil {
		return
	}

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#3B82F6")).
		Padding(0, 1)

	platformBadgeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#8B5CF6")).
		Padding(0, 1)

	metaLabelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#94A3B8"))

	metaValStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#F8FAFC"))

	sectionTitleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#60A5FA"))

	createdStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#10B981"))

	skippedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F59E0B"))

	stepNumberStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#6366F1"))

	stepTextStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#E2E8F0"))

	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintf(w, "%s %s\n",
		headerStyle.Render(" UMARU DEPLOY "),
		platformBadgeStyle.Render(strings.ToUpper(string(result.Platform))),
	)
	_, _ = fmt.Fprintf(w, "%s %s  %s %d  %s %s\n",
		metaLabelStyle.Render("App:"), metaValStyle.Render(result.AppName),
		metaLabelStyle.Render("Port:"), result.Port,
		metaLabelStyle.Render("Path:"), metaValStyle.Render(result.TargetDir),
	)
	_, _ = fmt.Fprintln(w)

	// Files table
	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#475569"))).
		Headers("DEPLOYMENT MANIFEST", "STATUS")

	for _, f := range result.Files {
		var actionStr string
		switch {
		case strings.HasPrefix(f.Action, "created") || strings.HasPrefix(f.Action, "create"):
			actionStr = createdStyle.Render("✔ " + f.Action)
		case strings.HasPrefix(f.Action, "overwritten") || strings.HasPrefix(f.Action, "overwrite"):
			actionStr = createdStyle.Render("✔ " + f.Action)
		default:
			actionStr = skippedStyle.Render("⚠ " + f.Action)
		}

		t.Row(f.Path, actionStr)
	}

	_, _ = fmt.Fprintln(w, t.Render())

	// Next Steps
	if len(result.Instructions) > 0 {
		_, _ = fmt.Fprintln(w)
		_, _ = fmt.Fprintln(w, sectionTitleStyle.Render("Next Steps & Deployment Commands:"))
		for i, inst := range result.Instructions {
			_, _ = fmt.Fprintf(w, "  %s %s\n",
				stepNumberStyle.Render(fmt.Sprintf("%d.", i+1)),
				stepTextStyle.Render(inst),
			)
		}
	}

	if dryRun {
		_, _ = fmt.Fprintln(w)
		dryStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FBBF24")).
			Bold(true)
		_, _ = fmt.Fprintln(w, dryStyle.Render("💡 Dry-run mode: no files were written to disk. Remove --dry-run to generate."))
	}

	_, _ = fmt.Fprintln(w)
}
