package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"umaru/internal/cleaner"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"
)

var (
	cleanDirFlag       string
	cleanRecursiveFlag bool
	cleanDryRunFlag    bool
	cleanForceFlag     bool
	cleanAllFlag       bool
	cleanJSONFlag      bool
)

var cleanCmd = &cobra.Command{
	Use:     "clean [path]",
	Aliases: []string{"sanitize", "purge"},
	Short:   "Reclaim disk space by cleaning build artifacts, caches, and dependency folders",
	GroupID: "util",
	Args:    cobra.MaximumNArgs(1),
	Example: `  umaru clean
  umaru clean --dry-run
  umaru clean -f
  umaru clean -r
  umaru clean ./apps/api
  umaru clean --all -f
  umaru clean --json`,
	Long: `Scan the target project directory to detect and safely delete build artifacts,
package dependencies, and cache folders that consume significant disk space:
  - Node.js   : node_modules, .next, .nuxt, .turbo, .astro, .svelte-kit, .cache
  - Build     : dist, build, out, coverage
  - Rust      : target
  - Python    : __pycache__, .pytest_cache, .mypy_cache, *.pyc (.venv with --all)
  - OS / Temp : tmp, .DS_Store, Thumbs.db`,
	RunE: func(cmd *cobra.Command, args []string) error {
		targetDir := "."
		if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
			targetDir = strings.TrimSpace(args[0])
		} else if cleanDirFlag != "" {
			targetDir = cleanDirFlag
		}

		opts := cleaner.ScanOptions{
			RootDir:    targetDir,
			Recursive:  cleanRecursiveFlag,
			IncludeAll: cleanAllFlag,
		}

		report, err := cleaner.Scan(opts)
		if err != nil {
			return err
		}

		// JSON Output handling
		if cleanJSONFlag {
			if cleanDryRunFlag || (!cleanForceFlag && !isTerminalStdin()) {
				data, err := json.MarshalIndent(report, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}

			result, err := cleaner.ExecuteClean(report)
			if err != nil {
				return err
			}

			output := map[string]interface{}{
				"report": report,
				"result": result,
			}
			data, err := json.MarshalIndent(output, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		}

		titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
		successStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#10B981"))
		warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B"))
		pathStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00D8F6"))
		sizeStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#A3E635"))

		fmt.Println()
		fmt.Printf("%s %s\n\n", titleStyle.Render("🧹 Umaru Project Cleaner:"), pathStyle.Render(report.RootDir))

		if len(report.Items) == 0 {
			fmt.Println(successStyle.Render("✨ Project is completely clean! No build artifacts or cache folders found."))
			fmt.Println()
			return nil
		}

		t := table.New().
			Border(lipgloss.RoundedBorder()).
			BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))).
			Headers("TARGET", "CATEGORY", "RECLAIMABLE SIZE")

		for _, item := range report.Items {
			relName := item.RelPath
			if item.IsDir {
				relName += "/"
			}
			t.Row(relName, string(item.Category), sizeStyle.Render(item.SizeDisplay))
		}

		fmt.Println(t)
		fmt.Printf("\n📦 Total Reclaimable Space: %s across %d targets\n\n", sizeStyle.Render(report.TotalDisplay), len(report.Items))

		if cleanDryRunFlag {
			fmt.Println(warnStyle.Render("💡 Dry-run mode: no files were deleted. Run 'umaru clean -f' to reclaim disk space."))
			fmt.Println()
			return nil
		}

		// Prompt user for confirmation if not forced
		if !cleanForceFlag {
			if !isTerminalStdin() {
				return fmt.Errorf("interactive confirmation unavailable in non-terminal mode. Use --force (-f) or --yes (-y)")
			}

			var confirm bool
			confirmForm := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title(fmt.Sprintf("Permanently delete these targets and reclaim %s?", report.TotalDisplay)).
						Value(&confirm),
				),
			)

			if err := confirmForm.Run(); err != nil {
				return err
			}

			if !confirm {
				fmt.Println(warnStyle.Render("Cleanup cancelled by user."))
				fmt.Println()
				return nil
			}
		}

		result, err := cleaner.ExecuteClean(report)
		if err != nil {
			return err
		}

		fmt.Println(successStyle.Render(fmt.Sprintf("✔ Successfully reclaimed %s of disk space across %d targets!", result.ReclaimedDisplay, result.DeletedCount)))
		fmt.Println()
		return nil
	},
}

func init() {
	cleanCmd.Flags().StringVarP(&cleanDirFlag, "dir", "d", ".", "Target project directory to clean")
	cleanCmd.Flags().BoolVarP(&cleanRecursiveFlag, "recursive", "r", false, "Scan recursively into subdirectories and monorepo packages")
	cleanCmd.Flags().BoolVar(&cleanDryRunFlag, "dry-run", false, "Simulate scan and show reclaimable space without deleting files")
	cleanCmd.Flags().BoolVarP(&cleanForceFlag, "force", "f", false, "Bypass interactive confirmation prompt")
	cleanCmd.Flags().BoolVarP(&cleanForceFlag, "yes", "y", false, "Automatic yes to confirmation prompt (alias for --force)")
	cleanCmd.Flags().BoolVar(&cleanAllFlag, "all", false, "Also remove virtual environments (.venv) and additional caches")
	cleanCmd.Flags().BoolVar(&cleanJSONFlag, "json", false, "Output scan report in JSON format")

	rootCmd.AddCommand(cleanCmd)
}
