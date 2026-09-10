package cmd

import (
	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script",
	Long: `Generate shell completion script for Umaru CLI.

To load completions:

Bash:
  $ source <(umaru completion bash)

  # To load completions for every new session, execute once:
  # Linux:
  $ umaru completion bash > /etc/bash_completion.d/umaru
  # macOS:
  $ umaru completion bash > $(brew --prefix)/etc/bash_completion.d/umaru

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ umaru completion zsh > "${fpath[1]}/_umaru"

Fish:
  $ umaru completion fish | source

  # To load completions for each session, execute once:
  $ umaru completion fish > ~/.config/fish/completions/umaru.fish

PowerShell:
  PS> umaru completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> umaru completion powershell > $PROFILE
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		out := cmd.OutOrStdout()
		switch args[0] {
		case "bash":
			_ = cmd.Root().GenBashCompletion(out)
		case "zsh":
			_ = cmd.Root().GenZshCompletion(out)
		case "fish":
			_ = cmd.Root().GenFishCompletion(out, true)
		case "powershell":
			_ = cmd.Root().GenPowerShellCompletionWithDesc(out)
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
