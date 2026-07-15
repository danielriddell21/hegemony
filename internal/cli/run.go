package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/danielriddell21/hegemony/internal/gui"
	"github.com/danielriddell21/hegemony/internal/sim"
	"github.com/danielriddell21/hegemony/internal/strategy"
)

func newRunCmd() *cobra.Command {
	var (
		width, height int
		ticks         int
		seed          uint64
		threshold     float64
		cellSize      int
		tps           int
		strategies    []string
	)

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Watch a single match in the GUI (build with -tags ebiten)",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			names := strategies
			if len(names) == 0 {
				names = strategy.Names()
			}
			if err := gui.Run(gui.Config{
				Width:          width,
				Height:         height,
				Seed:           seed,
				MaxTicks:       ticks,
				WinThreshold:   threshold,
				Params:         sim.DefaultParams(),
				Strategies:     names,
				CellSize:       cellSize,
				TicksPerSecond: tps,
			}); err != nil {
				return fmt.Errorf("run match: %w", err)
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&width, "width", 32, "grid width in cells")
	cmd.Flags().IntVar(&height, "height", 32, "grid height in cells")
	cmd.Flags().IntVar(&ticks, "ticks", 2000, "maximum ticks before the match ends")
	cmd.Flags().Uint64Var(&seed, "seed", 1, "match seed")
	cmd.Flags().Float64Var(&threshold, "threshold", 0.6, "territory fraction to win outright (0 disables)")
	cmd.Flags().IntVar(&cellSize, "cell-size", 18, "pixels per cell")
	cmd.Flags().IntVar(&tps, "tps", 12, "simulation ticks per second")
	cmd.Flags().StringSliceVar(&strategies, "strategies", nil, "strategies to enter (default: all registered)")
	return cmd
}
