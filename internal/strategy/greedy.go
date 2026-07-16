package strategy

import (
	"math/rand/v2"

	"github.com/danielriddell21/hegemony/internal/sim"
)

type greedy struct{}

func Greedy() sim.Strategy { return greedy{} }

func (greedy) Name() string { return "Greedy" }

func (greedy) Move(v sim.View, _ *rand.Rand) []sim.Action {
	front := v.Frontier()
	if len(front) == 0 {
		return nil
	}
	return captureByScore(v, front, func(p sim.Point) float64 {
		return -float64(v.At(p).Strength)
	})
}
