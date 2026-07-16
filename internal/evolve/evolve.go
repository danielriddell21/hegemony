package evolve

import (
	"math/rand/v2"

	"github.com/danielriddell21/hegemony/internal/sim"
	"github.com/danielriddell21/hegemony/internal/strategy"
)

type Config struct {
	Width       int
	Height      int
	Seeds       int
	Ticks       int
	Threshold   float64
	Params      sim.Params
	Opponents   []sim.Strategy
	Generations int
	Seed        uint64
}

type Result struct {
	Weights    strategy.Weights
	Fitness    float64
	Baseline   float64
	Generation int
}

// Run hill-climbs a weight vector with annealed Gaussian mutation, scoring each
// candidate by the mean territory it wins in a free-for-all against Opponents.
func Run(cfg Config) Result {
	rng := rand.New(rand.NewPCG(cfg.Seed, 0x5EED))
	best := strategy.Weights{Weak: 1, Open: 1, Influence: 0.5, EnemyDist: 0.5, Source: 0.5}
	baseline := fitness(cfg, best)
	bestFit := baseline
	bestGen := 0

	step := 1.0
	for g := 1; g <= cfg.Generations; g++ {
		cand := mutate(best, rng, step)
		if fit := fitness(cfg, cand); fit > bestFit {
			best, bestFit, bestGen = cand, fit, g
		}
		step *= 0.97
	}
	return Result{Weights: best, Fitness: bestFit, Baseline: baseline, Generation: bestGen}
}

func mutate(w strategy.Weights, rng *rand.Rand, step float64) strategy.Weights {
	return strategy.Weights{
		Weak:      w.Weak + rng.NormFloat64()*step,
		Open:      w.Open + rng.NormFloat64()*step,
		Influence: w.Influence + rng.NormFloat64()*step,
		EnemyDist: w.EnemyDist + rng.NormFloat64()*step,
		Source:    w.Source + rng.NormFloat64()*step,
	}
}

// fitness duels the candidate one-on-one against each opponent over several
// seeds and averages the territory share it holds. Duels give a clean gradient:
// in a crowded free-for-all one strategy tends to snowball and everyone else is
// eliminated, which flattens the signal to noise.
func fitness(cfg Config, w strategy.Weights) float64 {
	me := strategy.Weighted(w)
	spawns := sim.SpreadSpawns(cfg.Width, cfg.Height, 2)

	sum := 0.0
	n := 0
	for oi, opp := range cfg.Opponents {
		for s := range cfg.Seeds {
			res := sim.RunMatch(sim.MatchConfig{
				Width:        cfg.Width,
				Height:       cfg.Height,
				Seed:         cfg.Seed + uint64(oi*cfg.Seeds+s) + 1,
				MaxTicks:     cfg.Ticks,
				WinThreshold: cfg.Threshold,
				Params:       cfg.Params,
				Entrants: []sim.Entrant{
					{Strategy: me, Spawn: spawns[0]},
					{Strategy: opp, Spawn: spawns[1]},
				},
			})
			sum += meanShare(res.Share, 1)
			n++
		}
	}
	return sum / float64(n)
}

func meanShare(share [][]float64, id sim.FactionID) float64 {
	if len(share) == 0 {
		return 0
	}
	sum := 0.0
	for _, row := range share {
		sum += row[id]
	}
	return sum / float64(len(share))
}
