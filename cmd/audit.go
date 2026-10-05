package cmd

import (
	"fmt"
	"umaru/internal/audit"

	"github.com/spf13/cobra"
)

var (
	auditDirFlag       string
	auditJSONFlag      bool
	auditStrictFlag    bool
	auditNoNetworkFlag bool
)

var auditCmd = &cobra.Command{
	Use:     "audit [path]",
	Aliases: []string{"check", "inspect"},
	Short:   "Audit project health, security, configuration parity, and dependencies",
	GroupID: "util",
	Example: `  umaru audit
  umaru audit ./my-app
  umaru audit --strict
  umaru audit --json
  umaru audit --no-network`,
	Long: `Audit runs comprehensive diagnostics against an existing project directory.
It checks:
  • Environment variable parity (.env vs .env.example)
  • Secret leak vulnerabilities (.env included in .gitignore)
  • Missing dependencies (node_modules, .venv, go.sum)
  • Port collisions and availability
  • Overall project health score and prioritized actionable recommendations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		targetDir := auditDirFlag
		if len(args) > 0 && args[0] != "" {
			targetDir = args[0]
		}
		if targetDir == "" {
			targetDir = "."
		}

		opts := audit.AuditOptions{
			TargetDir:    targetDir,
			CheckNetwork: !auditNoNetworkFlag,
		}

		report, err := audit.RunAudit(opts)
		if err != nil {
			return fmt.Errorf("audit failed: %w", err)
		}

		if auditJSONFlag {
			data, err := report.ToJSON(true)
			if err != nil {
				return fmt.Errorf("failed serializing audit report to json: %w", err)
			}
			fmt.Println(string(data))
		} else {
			audit.RenderReport(report, verboseFlag)
		}

		if auditStrictFlag {
			hasCriticalFailures := false
			for _, check := range report.Checks {
				if check.Status == audit.StatusFail {
					hasCriticalFailures = true
					break
				}
			}
			if report.HealthScore < 80 || hasCriticalFailures {
				return fmt.Errorf("audit failed strict criteria: health score is %d/100 with critical issues", report.HealthScore)
			}
		}

		return nil
	},
}

func init() {
	auditCmd.Flags().StringVarP(&auditDirFlag, "dir", "d", ".", "Target project directory to audit")
	auditCmd.Flags().BoolVar(&auditJSONFlag, "json", false, "Output audit results in JSON format")
	auditCmd.Flags().BoolVar(&auditStrictFlag, "strict", false, "Exit with non-zero code if critical warnings or low health score (<80)")
	auditCmd.Flags().BoolVar(&auditNoNetworkFlag, "no-network", false, "Skip network port availability testing")

	rootCmd.AddCommand(auditCmd)
}
