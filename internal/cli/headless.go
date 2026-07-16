package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/danielriddell21/hegemony/internal/strategy"
	"github.com/danielriddell21/hegemony/internal/tournament"
)

func newHeadlessCmd() *cobra.Command {
	var (
		width, height int
		ticks, seeds  int
		baseSeed      uint64
		threshold     float64
		strategies    []string
	)

	cmd := &cobra.Command{
		Use:   "headless",
		Short: "Run a batch tournament across strategies and print a result table",
		Args:  cobra.NoArgs,
	}
	params := addParamFlags(cmd)
	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		names := strategies
		if len(names) == 0 {
			names = strategy.Names()
		}
		standings, err := tournament.Run(tournament.Config{
			Width:        width,
			Height:       height,
			Seeds:        seeds,
			BaseSeed:     baseSeed,
			MaxTicks:     ticks,
			WinThreshold: threshold,
			Params:       params.params(),
			Strategies:   names,
		})
		if err != nil {
			return fmt.Errorf("run tournament: %w", err)
		}
		fmt.Fprint(cmd.OutOrStdout(), tournament.Table(standings))
		return nil
	}

	cmd.Flags().IntVar(&width, "width", 24, "grid width in cells")
	cmd.Flags().IntVar(&height, "height", 24, "grid height in cells")
	cmd.Flags().IntVar(&ticks, "ticks", 400, "maximum ticks per match")
	cmd.Flags().IntVar(&seeds, "seeds", 50, "number of seeded matches to run")
	cmd.Flags().Uint64Var(&baseSeed, "seed", 1, "base seed; match i uses seed+i")
	cmd.Flags().Float64Var(&threshold, "threshold", 0.6, "territory fraction to win outright (0 disables)")
	cmd.Flags().StringSliceVar(&strategies, "strategies", nil, "strategies to enter (default: all registered)")
	return cmd
}
