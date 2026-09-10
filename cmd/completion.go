package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	completionInstallFlag bool
	completionTestHomeDir string // used for isolated testing
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script or install it automatically",
	Long: `Generate shell completion script for Umaru CLI, or install it to your shell profile.

To load completions automatically in your current shell:
  $ umaru completion --install

Or specify your target shell:
  $ umaru completion zsh --install
  $ umaru completion powershell --install

Manual loading:

Bash:
  $ source <(umaru completion bash)
  # Persistent:
  $ umaru completion bash > /etc/bash_completion.d/umaru

Zsh:
  $ eval "$(umaru completion zsh)"
  # Persistent:
  $ echo 'eval "$(umaru completion zsh)"' >> ~/.zshrc

Fish:
  $ umaru completion fish | source
  # Persistent:
  $ umaru completion fish > ~/.config/fish/completions/umaru.fish

PowerShell:
  PS> umaru completion powershell | Out-String | Invoke-Expression
  # Persistent (safe install):
  PS> umaru completion --install
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args: func(cmd *cobra.Command, args []string) error {
		installFlag, _ := cmd.Flags().GetBool("install")
		if installFlag {
			if len(args) == 0 {
				return nil
			}
			if len(args) == 1 {
				for _, valid := range []string{"bash", "zsh", "fish", "powershell"} {
					if args[0] == valid {
						return nil
					}
				}
				return fmt.Errorf("invalid shell %q, must be bash, zsh, fish, or powershell", args[0])
			}
			return fmt.Errorf("accepts at most 1 arg (%d given)", len(args))
		}

		if len(args) != 1 {
			return fmt.Errorf("accepts 1 arg (%d given). Specify bash, zsh, fish, or powershell, or use --install", len(args))
		}
		for _, valid := range []string{"bash", "zsh", "fish", "powershell"} {
			if args[0] == valid {
				return nil
			}
		}
		return fmt.Errorf("invalid shell %q, must be bash, zsh, fish, or powershell", args[0])
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		installFlag, _ := cmd.Flags().GetBool("install")
		if installFlag {
			targetShell := ""
			if len(args) > 0 {
				targetShell = args[0]
			} else {
				targetShell = detectShell()
			}
			return installCompletion(cmd, targetShell)
		}

		out := cmd.OutOrStdout()
		switch args[0] {
		case "bash":
			return cmd.Root().GenBashCompletion(out)
		case "zsh":
			return cmd.Root().GenZshCompletion(out)
		case "fish":
			return cmd.Root().GenFishCompletion(out, true)
		case "powershell":
			return cmd.Root().GenPowerShellCompletionWithDesc(out)
		}
		return nil
	},
}

func init() {
	completionCmd.Flags().BoolVarP(&completionInstallFlag, "install", "i", false, "Automatically install completion to your shell profile")
	rootCmd.AddCommand(completionCmd)
}

func detectShell() string {
	shellEnv := os.Getenv("SHELL")
	if shellEnv != "" {
		base := strings.ToLower(filepath.Base(shellEnv))
		switch {
		case strings.Contains(base, "zsh"):
			return "zsh"
		case strings.Contains(base, "bash"):
			return "bash"
		case strings.Contains(base, "fish"):
			return "fish"
		}
	}

	if runtime.GOOS == "windows" {
		return "powershell"
	}

	return "bash"
}

func getHomeDir() (string, error) {
	if completionTestHomeDir != "" {
		return completionTestHomeDir, nil
	}
	return os.UserHomeDir()
}

func getPowerShellProfilePath(home string) string {
	if completionTestHomeDir == "" && runtime.GOOS == "windows" {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		out, err := exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command", "Write-Output $PROFILE").Output()
		if err == nil {
			p := strings.TrimSpace(string(out))
			if p != "" {
				return p
			}
		}
	}
	return filepath.Join(home, "Documents", "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1")
}

func installCompletion(cmd *cobra.Command, targetShell string) error {
	home, err := getHomeDir()
	if err != nil {
		return fmt.Errorf("could not determine user home directory: %w", err)
	}

	out := cmd.OutOrStdout()

	switch targetShell {
	case "powershell":
		profilePath := getPowerShellProfilePath(home)
		snippet := "\n# Umaru CLI autocompletion\nif (Get-Command umaru -ErrorAction SilentlyContinue) {\n    umaru completion powershell | Out-String | Invoke-Expression\n}\n"
		return appendToConfigFile(out, profilePath, snippet, "umaru completion powershell")

	case "zsh":
		zshrcPath := filepath.Join(home, ".zshrc")
		snippet := "\n# Umaru CLI autocompletion\nif command -v umaru &> /dev/null; then\n    eval \"$(umaru completion zsh)\"\nfi\n"
		return appendToConfigFile(out, zshrcPath, snippet, "umaru completion zsh")

	case "bash":
		bashrcPath := filepath.Join(home, ".bashrc")
		snippet := "\n# Umaru CLI autocompletion\nif command -v umaru &> /dev/null; then\n    eval \"$(umaru completion bash)\"\nfi\n"
		return appendToConfigFile(out, bashrcPath, snippet, "umaru completion bash")

	case "fish":
		fishDir := filepath.Join(home, ".config", "fish", "completions")
		if err := os.MkdirAll(fishDir, 0755); err != nil {
			return fmt.Errorf("failed to create fish completions directory: %w", err)
		}
		fishFile := filepath.Join(fishDir, "umaru.fish")
		f, err := os.Create(fishFile)
		if err != nil {
			return fmt.Errorf("failed to create fish completion file: %w", err)
		}
		defer f.Close()
		if err := cmd.Root().GenFishCompletion(f, true); err != nil {
			return fmt.Errorf("failed to write fish completion: %w", err)
		}
		fmt.Fprintf(out, "✔ Successfully installed Fish completion script to:\n  %s\n", fishFile)
		return nil

	default:
		return fmt.Errorf("unsupported shell: %s", targetShell)
	}
}

func appendToConfigFile(out io.Writer, filePath, snippet, checkSignature string) error {
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", filePath, err)
	}

	if content, err := os.ReadFile(filePath); err == nil {
		if strings.Contains(string(content), checkSignature) {
			fmt.Fprintf(out, "ℹ️ Umaru completion is already installed in:\n  %s\n", filePath)
			return nil
		}
	}

	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", filePath, err)
	}
	defer f.Close()

	if _, err := f.WriteString(snippet); err != nil {
		return fmt.Errorf("failed to write to %s: %w", filePath, err)
	}

	fmt.Fprintf(out, "✔ Successfully installed Umaru autocompletion to:\n  %s\n\n💡 Restart your terminal or reload your shell profile to activate tab-completions!\n", filePath)
	return nil
}
