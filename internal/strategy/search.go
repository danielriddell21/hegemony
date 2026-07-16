package strategy

import (
	"math/rand/v2"

	"github.com/danielriddell21/hegemony/internal/sim"
)

const (
	searchDepth = 8
	searchSeed  = 1
)

type search struct {
	panel  []sim.Strategy
	rollFn func() sim.Strategy
}

func Search() sim.Strategy {
	return search{
		panel:  []sim.Strategy{Greedy(), Frontier(), Influence(), Bulwark()},
		rollFn: Bulwark,
	}
}

func (search) Name() string { return "Search" }

func (s search) Move(v sim.View, rng *rand.Rand) []sim.Action {
	if len(v.Frontier()) == 0 {
		return nil
	}
	cells := snapshotCells(v)
	me := v.Faction()
	params := v.Params()
	params.Jitter = 0

	var bestCand sim.Strategy
	bestScore := -1
	for _, cand := range s.panel {
		if score := s.project(cells, params, me, v.Width(), v.Height(), cand); score > bestScore {
			bestScore, bestCand = score, cand
		}
	}
	if bestCand == nil {
		return nil
	}
	return bestCand.Move(v, rng)
}

// project forks the position, commits our faction to cand while every opponent
// plays a fixed policy, rolls the game forward a few ticks, and reports the
// territory we end up holding — so Search picks the plan that projects best
// rather than the move that looks best for a single tick.
func (s search) project(cells []sim.Cell, params sim.Params, me sim.FactionID, w, h int, cand sim.Strategy) int {
	policies := make(map[sim.FactionID]sim.Strategy)
	for _, c := range cells {
		if c.Owner != sim.Neutral {
			policies[c.Owner] = s.rollFn()
		}
	}
	policies[me] = cand

	world := sim.NewSnapshotWorld(sim.SnapshotConfig{
		Width:    w,
		Height:   h,
		Cells:    cells,
		Params:   params,
		Seed:     searchSeed,
		Policies: policies,
	})
	for range searchDepth {
		world.Tick()
	}
	return world.Territory(me)
}

func snapshotCells(v sim.View) []sim.Cell {
	cells := make([]sim.Cell, v.Width()*v.Height())
	for y := range v.Height() {
		for x := range v.Width() {
			cells[y*v.Width()+x] = v.At(sim.Point{X: x, Y: y})
		}
	}
	return cells
}
