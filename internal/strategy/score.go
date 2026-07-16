package strategy

import (
	"cmp"
	"slices"

	"github.com/danielriddell21/hegemony/internal/sim"
)

type scoredCell struct {
	p     sim.Point
	score float64
}

func captureByScore(v sim.View, cells []sim.Point, score func(sim.Point) float64) []sim.Action {
	scored := make([]scoredCell, len(cells))
	for i, p := range cells {
		scored[i] = scoredCell{p: p, score: score(p)}
	}
	slices.SortFunc(scored, func(a, b scoredCell) int {
		if c := cmp.Compare(b.score, a.score); c != 0 {
			return c
		}
		if c := cmp.Compare(a.p.Y, b.p.Y); c != 0 {
			return c
		}
		return cmp.Compare(a.p.X, b.p.X)
	})

	budget := v.Budget()
	actions := make([]sim.Action, 0, len(scored))
	for _, sc := range scored {
		c := v.At(sc.p)
		cost := captureCost(c.Strength)
		if cost > budget {
			continue
		}
		actions = append(actions, sim.Action{Kind: captureKind(c.Owner), Cell: sc.p, Amount: cost})
		budget -= cost
	}
	return actions
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
