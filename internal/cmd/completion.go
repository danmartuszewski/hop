package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script",
	Long: `Generate shell completion script for hop.

To load completions:

Bash:
  $ source <(hop completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ hop completion bash > /etc/bash_completion.d/hop
  # macOS:
  $ hop completion bash > $(brew --prefix)/etc/bash_completion.d/hop

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ hop completion zsh > "${fpath[1]}/_hop"

  # You may need to start a new shell for this setup to take effect.

Fish:
  $ hop completion fish | source

  # To load completions for each session, execute once:
  $ hop completion fish > ~/.config/fish/completions/hop.fish

PowerShell:
  PS> hop completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> hop completion powershell > hop.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			return cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			return cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)

	// Register dynamic completion for connection IDs
	rootCmd.RegisterFlagCompletionFunc("config", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveDefault
	})
}

// getConnectionCompletions returns connection IDs for shell completion
func getConnectionCompletions(toComplete string) ([]string, cobra.ShellCompDirective) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	for _, conn := range cfg.Connections {
		completions = append(completions, conn.ID)
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}
