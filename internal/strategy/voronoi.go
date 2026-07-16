package strategy

import (
	"math/rand/v2"

	"github.com/danielriddell21/hegemony/internal/sim"
)

type voronoi struct{}

func Voronoi() sim.Strategy { return voronoi{} }

func (voronoi) Name() string { return "Voronoi" }

func (voronoi) Move(v sim.View, _ *rand.Rand) []sim.Action {
	front := v.Frontier()
	if len(front) == 0 {
		return nil
	}

	_, enemies := scanOwners(v)
	if len(enemies) == 0 {
		return captureByScore(v, front, func(p sim.Point) float64 { return roi(v, p) })
	}

	// Claim the hinterland that is unambiguously ours — cells far from any
	// enemy — before contesting the border, locking in the larger region.
	return captureByScore(v, front, func(p sim.Point) float64 {
		return float64(minDistance(p, enemies))
	})
}
