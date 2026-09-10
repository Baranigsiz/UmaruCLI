package cmd

import (
	"umaru/internal/doctor"

	"github.com/spf13/cobra"
)

var verboseDoctor bool

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Inspect developer environment and verify template readiness",
	Long: `Doctor checks your system for required tools, programming runtimes (Go, Node, Python, Rust),
package managers, containers (Docker & Compose), and reports readiness across all starter templates.`,
	Run: func(cmd *cobra.Command, args []string) {
		report := doctor.RunDiagnostics(Version)
		doctor.RenderReport(report, verboseDoctor)
	},
}

func init() {
	doctorCmd.Flags().BoolVarP(&verboseDoctor, "verbose", "v", false, "Show binary paths and verbose diagnostic information")
	rootCmd.AddCommand(doctorCmd)
}
