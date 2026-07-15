package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newCompletionCmd() *cobra.Command {
	return &cobra.Command{
		Use:                   "completion [bash|zsh|fish|powershell]",
		Short:                 "Generate the autocompletion script for the specified shell",
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			switch args[0] {
			case "bash":
				err = cmd.Root().GenBashCompletionV2(os.Stdout, true)
			case "zsh":
				err = cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				err = cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				err = cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
			if err != nil {
				return fmt.Errorf("generate %s completion: %w", args[0], err)
			}
			return nil
		},
	}
}
