package strategy

import (
	"math/rand/v2"

	"github.com/danielriddell21/hegemony/internal/sim"
)

type frontier struct{}

func Frontier() sim.Strategy { return frontier{} }

func (frontier) Name() string { return "Frontier" }

func (frontier) Move(v sim.View, _ *rand.Rand) []sim.Action {
	front := v.Frontier()
	if len(front) == 0 {
		return nil
	}
	return captureByScore(v, front, func(p sim.Point) float64 { return roi(v, p) })
}

func roi(v sim.View, p sim.Point) float64 {
	openness := 1
	for _, q := range v.Neighbors(p) {
		if v.At(q).Owner == sim.Neutral {
			openness++
		}
	}
	return float64(openness) / float64(captureCost(v.At(p).Strength))
}
