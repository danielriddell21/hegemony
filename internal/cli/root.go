package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func Execute(version string) error {
	if err := newRootCmd(version).Execute(); err != nil {
		return fmt.Errorf("hegemony: %w", err)
	}
	return nil
}

func newRootCmd(version string) *cobra.Command {
	root := &cobra.Command{
		Use:   "hegemony",
		Short: "Territory-war simulation where competing algorithms fight to control a shared map",
		Long: "hegemony pits pluggable strategies against each other on a shared grid: " +
			"factions expand, attack, and defend to claim territory. Run a batch tournament " +
			"headlessly, or watch a single match in the GUI (build with -tags ebiten).",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(newHeadlessCmd(), newRunCmd(), newEvolveCmd(), newCompletionCmd())
	return root
}
