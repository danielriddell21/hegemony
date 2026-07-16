package strategy

import (
	"cmp"
	"slices"

	"github.com/danielriddell21/hegemony/internal/sim"
)

func captureByScore(v sim.View, cells []sim.Point, score func(sim.Point) float64) []sim.Action {
	pl := newPlan(v)
	for _, target := range sortedByScore(cells, score) {
		pl.capture(target)
	}
	return pl.result()
}

func sortedByScore(cells []sim.Point, score func(sim.Point) float64) []sim.Point {
	type scored struct {
		p     sim.Point
		score float64
	}
	ranked := make([]scored, len(cells))
	for i, p := range cells {
		ranked[i] = scored{p: p, score: score(p)}
	}
	slices.SortFunc(ranked, func(a, b scored) int {
		if c := cmp.Compare(b.score, a.score); c != 0 {
			return c
		}
		if c := cmp.Compare(a.p.Y, b.p.Y); c != 0 {
			return c
		}
		return cmp.Compare(a.p.X, b.p.X)
	})
	out := make([]sim.Point, len(ranked))
	for i, r := range ranked {
		out[i] = r.p
	}
	return out
}

func captureCost(strength int) int {
	return strength + strength/4 + 1
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
