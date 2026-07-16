package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/danielriddell21/hegemony/internal/evolve"
	"github.com/danielriddell21/hegemony/internal/sim"
	"github.com/danielriddell21/hegemony/internal/strategy"
)

var defaultOpponents = []string{"Greedy", "Blob", "Frontier", "Voronoi", "Influence"}

func resolveOpponents(names []string) ([]sim.Strategy, error) {
	out := make([]sim.Strategy, 0, len(names))
	for _, name := range names {
		s, ok := strategy.New(name)
		if !ok {
			return nil, fmt.Errorf("unknown strategy %q", name)
		}
		out = append(out, s)
	}
	return out, nil
}

func newEvolveCmd() *cobra.Command {
	var (
		width, height int
		seeds, ticks  int
		generations   int
		threshold     float64
		baseSeed      uint64
		opponents     []string
	)

	cmd := &cobra.Command{
		Use:   "evolve",
		Short: "Search weights for the Evolved strategy and print the best vector",
		Args:  cobra.NoArgs,
	}
	params := addParamFlags(cmd)
	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		names := opponents
		if len(names) == 0 {
			names = defaultOpponents
		}
		resolved, err := resolveOpponents(names)
		if err != nil {
			return err
		}
		res := evolve.Run(evolve.Config{
			Width:       width,
			Height:      height,
			Seeds:       seeds,
			Ticks:       ticks,
			Threshold:   threshold,
			Params:      params.params(),
			Opponents:   resolved,
			Generations: generations,
			Seed:        baseSeed,
		})
		out := cmd.OutOrStdout()
		fmt.Fprintf(out, "opponents: %v\n", names)
		fmt.Fprintf(out, "baseline territory: %.3f\n", res.Baseline)
		fmt.Fprintf(out, "best territory:     %.3f  (found at generation %d)\n", res.Fitness, res.Generation)
		fmt.Fprintf(out, "\nWeights{Weak: %.3f, Open: %.3f, Influence: %.3f, EnemyDist: %.3f, Source: %.3f}\n",
			res.Weights.Weak, res.Weights.Open, res.Weights.Influence, res.Weights.EnemyDist, res.Weights.Source)
		return nil
	}

	cmd.Flags().IntVar(&width, "width", 24, "grid width in cells")
	cmd.Flags().IntVar(&height, "height", 24, "grid height in cells")
	cmd.Flags().IntVar(&seeds, "seeds", 24, "matches per fitness evaluation")
	cmd.Flags().IntVar(&ticks, "ticks", 400, "maximum ticks per match")
	cmd.Flags().IntVar(&generations, "generations", 60, "hill-climb generations")
	cmd.Flags().Float64Var(&threshold, "threshold", 0, "territory fraction to win outright (0 disables)")
	cmd.Flags().Uint64Var(&baseSeed, "seed", 1, "evolution seed")
	cmd.Flags().StringSliceVar(&opponents, "opponents", nil, "strategies to evolve against (default: a fixed panel)")
	return cmd
}
