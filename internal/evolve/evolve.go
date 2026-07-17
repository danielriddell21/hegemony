package evolve

import (
	"github.com/danielriddell21/galapagos/pkg/ga"

	"github.com/danielriddell21/hegemony/internal/sim"
	"github.com/danielriddell21/hegemony/internal/strategy"
)

const (
	genomeLen         = 5
	defaultPopulation = 40
)

var incumbent = strategy.Weights{Weak: 1, Open: 1, Influence: 0.5, EnemyDist: 0.5, Source: 0.5}

type Config struct {
	Width       int
	Height      int
	Seeds       int
	Ticks       int
	Threshold   float64
	Params      sim.Params
	Opponents   []sim.Strategy
	Population  int
	Generations int
	Seed        uint64
}

type Result struct {
	Weights    strategy.Weights
	Fitness    float64
	Baseline   float64
	Generation int
}

func Run(cfg Config) Result {
	population := cfg.Population
	if population <= 0 {
		population = defaultPopulation
	}
	baseline := fitness(cfg, incumbent)
	score := func(g []float64) float64 { return fitness(cfg, weightsOf(g)) }

	res := ga.Run(ga.Config{
		PopulationSize: population,
		EliteFraction:  0.1,
		MutationRate:   0.3,
		MutationStd:    0.5,
		Seed:           int64(cfg.Seed),
	}, genomeLen, cfg.Generations, score, genomeOf(incumbent))

	return Result{
		Weights:    weightsOf(res.Best),
		Fitness:    res.Fitness,
		Baseline:   baseline,
		Generation: res.Generations,
	}
}

func genomeOf(w strategy.Weights) []float64 {
	return []float64{w.Weak, w.Open, w.Influence, w.EnemyDist, w.Source}
}

func weightsOf(g []float64) strategy.Weights {
	return strategy.Weights{Weak: g[0], Open: g[1], Influence: g[2], EnemyDist: g[3], Source: g[4]}
}

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
