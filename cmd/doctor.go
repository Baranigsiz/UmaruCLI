package cmd

import (
	"encoding/json"
	"fmt"
	"umaru/internal/doctor"

	"github.com/spf13/cobra"
)

var (
	verboseDoctor  bool
	doctorJSONFlag bool
)

var doctorCmd = &cobra.Command{
	Use:     "doctor",
	Short:   "Inspect developer environment and verify template readiness",
	GroupID: "util",
	Long: `Doctor checks your system for required tools, programming runtimes (Go, Node, Python, Rust),
package managers, containers (Docker & Compose), and reports readiness across all starter templates.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		report := doctor.RunDiagnostics(Version)
		if doctorJSONFlag {
			data, err := json.MarshalIndent(report, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to serialize doctor report to json: %w", err)
			}
			fmt.Println(string(data))
			return nil
		}
		doctor.RenderReport(report, verboseDoctor)
		return nil
	},
}

func init() {
	doctorCmd.Flags().BoolVarP(&verboseDoctor, "verbose", "v", false, "Show binary paths and verbose diagnostic information")
	doctorCmd.Flags().BoolVar(&doctorJSONFlag, "json", false, "Output doctor diagnostics in JSON format")
	rootCmd.AddCommand(doctorCmd)
}
