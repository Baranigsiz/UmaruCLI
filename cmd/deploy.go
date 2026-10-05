package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"umaru/internal/deploy"

	"github.com/charmbracelet/huh"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

var (
	deployPlatformFlag string
	deployDirFlag      string
	deployAppNameFlag  string
	deployPortFlag     int
	deployForceFlag    bool
	deployDryRunFlag   bool
	deployJSONFlag     bool
)

var deployCmd = &cobra.Command{
	Use:     "deploy [platform]",
	Aliases: []string{"ship", "publish"},
	Short:   "Generate production cloud deployment configuration and instructions",
	GroupID: "core",
	Example: `  umaru deploy
  umaru deploy fly
  umaru deploy railway
  umaru deploy render
  umaru deploy docker
  umaru deploy fly --app-name my-cool-app --port 8080
  umaru deploy render --dry-run
  umaru deploy railway --json`,
	Long: `Inspects your project and generates production-ready cloud deployment manifests:
  • fly      : Fly.io configuration (fly.toml, machine scaling & auto-stop)
  • railway  : Railway Blueprint (railway.json, Nixpacks builder & healthcheck)
  • render   : Render Blueprint (render.yaml, native runtime & environment variables)
  • docker   : Ultra-minimal multi-stage production container (Dockerfile.prod)`,
	ValidArgs: []string{"fly", "railway", "render", "docker"},
	Args:      cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		targetDir := deployDirFlag
		if targetDir == "" {
			targetDir = "."
		}

		var platform string
		if len(args) > 0 && args[0] != "" {
			platform = args[0]
		} else if deployPlatformFlag != "" {
			platform = deployPlatformFlag
		}

		if platform == "" {
			if !isatty.IsTerminal(os.Stdin.Fd()) && !isatty.IsCygwinTerminal(os.Stdin.Fd()) {
				return errors.New("platform required in non-interactive mode. Specify a platform (e.g. 'umaru deploy fly') or use --platform")
			}

			// Interactive selection with huh
			selectPrompt := huh.NewSelect[string]().
				Title("Select a Cloud Deployment Target").
				Description("Umaru will generate production manifests and CLI deployment instructions").
				Options(
					huh.NewOption("🎈 Fly.io (fly.toml — Global distributed apps & machines)", "fly"),
					huh.NewOption("🚂 Railway (railway.json — Instant cloud container hosting)", "railway"),
					huh.NewOption("🟣 Render (render.yaml — Managed infrastructure Blueprint)", "render"),
					huh.NewOption("🐳 Docker Production (Dockerfile.prod — Multi-stage minimal container)", "docker"),
				).
				Value(&platform)

			if err := selectPrompt.Run(); err != nil {
				return err
			}
		}

		opts := deploy.DeployOptions{
			TargetDir: targetDir,
			Platform:  deploy.TargetPlatform(strings.ToLower(strings.TrimSpace(platform))),
			AppName:   deployAppNameFlag,
			Port:      deployPortFlag,
			Force:     deployForceFlag,
			DryRun:    deployDryRunFlag,
		}

		result, err := deploy.GenerateDeployment(opts)
		if err != nil {
			return fmt.Errorf("deploy generation failed: %w", err)
		}

		if deployJSONFlag {
			data, err := result.ToJSON(true)
			if err != nil {
				return fmt.Errorf("failed serializing deploy result to json: %w", err)
			}
			fmt.Println(string(data))
			return nil
		}

		deploy.RenderResult(result, deployDryRunFlag)
		return nil
	},
}

func init() {
	deployCmd.Flags().StringVarP(&deployPlatformFlag, "platform", "p", "", "Target cloud platform (fly, railway, render, docker)")
	deployCmd.Flags().StringVarP(&deployDirFlag, "dir", "d", ".", "Target project directory to deploy")
	deployCmd.Flags().StringVar(&deployAppNameFlag, "app-name", "", "Custom application name for cloud services")
	deployCmd.Flags().IntVar(&deployPortFlag, "port", 0, "Internal HTTP port (default is detected from project stack)")
	deployCmd.Flags().BoolVarP(&deployForceFlag, "force", "f", false, "Overwrite existing deployment manifests")
	deployCmd.Flags().BoolVar(&deployDryRunFlag, "dry-run", false, "Preview generated manifests and instructions without writing to disk")
	deployCmd.Flags().BoolVar(&deployJSONFlag, "json", false, "Output deployment configuration and steps in JSON format")

	rootCmd.AddCommand(deployCmd)
}
