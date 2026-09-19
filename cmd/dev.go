package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"umaru/internal/devrunner"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	devDirFlag        string
	devPortFlag       string
	devHostFlag       string
	devPkgManagerFlag string
	devDryRunFlag     bool
	devJSONFlag       bool
)

var devCmd = &cobra.Command{
	Use:     "dev [path]",
	Aliases: []string{"run", "start"},
	Short:   "Zero-config development runner for Go, Node, Python, and Rust projects",
	GroupID: "core",
	Args:    cobra.MaximumNArgs(1),
	Example: `  umaru dev
  umaru dev --dry-run
  umaru dev -p 8080
  umaru dev --pm pnpm
  umaru dev ./my-project
  umaru run
  umaru start`,
	Long: `Automatically inspects the target project, resolves the language, framework,
package manager, and entrypoint, and launches the development server with live output:
  - Go     : go run cmd/api/main.go or cmd/web/main.go
  - Node   : pnpm dev / bun run dev / npm run dev (respects lockfiles)
  - Python : uvicorn app.main:app --reload
  - Rust   : cargo run
  - Docker : docker-compose up --build (monorepos)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		targetDir := "."
		if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
			targetDir = strings.TrimSpace(args[0])
		} else if devDirFlag != "" {
			targetDir = devDirFlag
		}

		opts := devrunner.DevOptions{
			TargetDir:      targetDir,
			Port:           devPortFlag,
			Host:           devHostFlag,
			PackageManager: devPkgManagerFlag,
		}

		cfg, err := devrunner.DetectDevCommand(opts)
		if err != nil {
			return err
		}

		if devJSONFlag {
			data, err := json.MarshalIndent(cfg, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		}

		titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
		successStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#10B981"))
		cmdStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00D8F6"))
		pathStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#94A3B8")).Italic(true)
		tagStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#A3E635"))

		fmt.Println()
		fmt.Printf("%s  %s %s\n", titleStyle.Render("⚡ Umaru Dev Runner:"), tagStyle.Render("["+strings.ToUpper(cfg.Language)+"]"), pathStyle.Render(cfg.TargetDir))
		fmt.Printf("🚀 Command: %s\n", cmdStyle.Render(cfg.CommandLine))
		if len(cfg.Env) > 0 {
			fmt.Printf("🔧 Environment: %s\n", strings.Join(cfg.Env, ", "))
		}
		fmt.Println()

		if devDryRunFlag {
			fmt.Println(successStyle.Render("💡 Dry-run mode: command resolved successfully. Remove --dry-run to start server."))
			fmt.Println()
			return nil
		}

		return devrunner.Run(cmd.Context(), cfg)
	},
}

func init() {
	devCmd.Flags().StringVarP(&devDirFlag, "dir", "d", ".", "Target project directory")
	devCmd.Flags().StringVarP(&devPortFlag, "port", "p", "", "Port to run the application on (sets PORT=...)")
	devCmd.Flags().StringVar(&devHostFlag, "host", "", "Host to bind the server to (sets HOST=...)")
	devCmd.Flags().StringVar(&devPkgManagerFlag, "pm", "", "Package manager override for Node projects (npm, pnpm, yarn, bun)")
	devCmd.Flags().BoolVar(&devDryRunFlag, "dry-run", false, "Preview detected command and environment without launching")
	devCmd.Flags().BoolVar(&devJSONFlag, "json", false, "Output resolved dev configuration as JSON")

	rootCmd.AddCommand(devCmd)
}
