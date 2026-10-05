package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"umaru/internal/testrunner"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	testDirFlag        string
	testVerboseFlag    bool
	testWatchFlag      bool
	testCoverageFlag   bool
	testFilterFlag     string
	testRaceFlag       bool
	testPkgManagerFlag string
	testDryRunFlag     bool
	testJSONFlag       bool
)

var testCmd = &cobra.Command{
	Use:     "test [path]",
	Aliases: []string{"t", "check-tests"},
	Short:   "Zero-config universal test runner for Go, Node, Python, and Rust projects",
	GroupID: "core",
	Args:    cobra.MaximumNArgs(1),
	Example: `  umaru test
  umaru test --watch
  umaru test --coverage
  umaru test --filter TestAuth
  umaru test --verbose
  umaru test --dry-run
  umaru test ./my-project`,
	Long: `Automatically inspects the target project, determines the language, testing framework,
and package manager, and executes tests with real-time streaming output:
  - Go     : go test -v ./... (with optional -race, -coverprofile, -run)
  - Node   : pnpm test / bun test / npm test (respects package.json test scripts)
  - Python : pytest -v (or python -m unittest if pytest is not installed)
  - Rust   : cargo test`,
	RunE: func(cmd *cobra.Command, args []string) error {
		targetDir := "."
		if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
			targetDir = strings.TrimSpace(args[0])
		} else if testDirFlag != "" {
			targetDir = testDirFlag
		}

		opts := testrunner.TestOptions{
			TargetDir:      targetDir,
			Verbose:        testVerboseFlag,
			Watch:          testWatchFlag,
			Coverage:       testCoverageFlag,
			Filter:         testFilterFlag,
			Race:           testRaceFlag,
			PackageManager: testPkgManagerFlag,
		}

		cfg, err := testrunner.DetectTestCommand(opts)
		if err != nil {
			return err
		}

		if testJSONFlag {
			data, err := json.MarshalIndent(cfg, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		}

		titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#EC4899"))
		successStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#10B981"))
		cmdStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#38BDF8"))
		pathStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#94A3B8")).Italic(true)
		tagStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#A78BFA"))

		fmt.Println()
		fmt.Printf("%s  %s %s\n", titleStyle.Render("🧪 Umaru Test Runner:"), tagStyle.Render("["+strings.ToUpper(cfg.Language)+"]"), pathStyle.Render(cfg.TargetDir))
		fmt.Printf("🚀 Command: %s\n", cmdStyle.Render(cfg.CommandLine))
		if len(cfg.Env) > 0 {
			fmt.Printf("🔧 Environment: %s\n", strings.Join(cfg.Env, ", "))
		}
		fmt.Println()

		if testDryRunFlag {
			fmt.Println(successStyle.Render("💡 Dry-run mode: test command resolved successfully. Remove --dry-run to execute tests."))
			fmt.Println()
			return nil
		}

		return testrunner.Run(cmd.Context(), cfg)
	},
}

func init() {
	testCmd.Flags().StringVarP(&testDirFlag, "dir", "d", ".", "Target project directory to test")
	testCmd.Flags().BoolVarP(&testVerboseFlag, "verbose", "v", false, "Run tests with verbose output")
	testCmd.Flags().BoolVarP(&testWatchFlag, "watch", "w", false, "Run tests in interactive watch / live mode")
	testCmd.Flags().BoolVarP(&testCoverageFlag, "coverage", "c", false, "Generate test code coverage report")
	testCmd.Flags().StringVarP(&testFilterFlag, "filter", "f", "", "Filter tests by name or pattern")
	testCmd.Flags().BoolVar(&testRaceFlag, "race", false, "Enable data race detection (Go projects)")
	testCmd.Flags().StringVar(&testPkgManagerFlag, "pm", "", "Override package manager (npm, pnpm, yarn, bun)")
	testCmd.Flags().BoolVar(&testDryRunFlag, "dry-run", false, "Print resolved test command without executing")
	testCmd.Flags().BoolVar(&testJSONFlag, "json", false, "Output resolved test configuration in JSON format")

	rootCmd.AddCommand(testCmd)
}
