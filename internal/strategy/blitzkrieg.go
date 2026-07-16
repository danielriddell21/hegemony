package strategy

import (
	"math/rand/v2"

	"github.com/danielriddell21/hegemony/internal/sim"
)

type blitzkrieg struct{}

func Blitzkrieg() sim.Strategy { return blitzkrieg{} }

func (blitzkrieg) Name() string { return "Blitzkrieg" }

func (blitzkrieg) Move(v sim.View, _ *rand.Rand) []sim.Action {
	front := v.Frontier()
	if len(front) == 0 {
		return nil
	}

	pl := newPlan(v)
	// Punch where we are strongest: pour a whole spearhead cell into the target
	// in front of it, concentrating force on one axis instead of spreading it.
	for _, target := range sortedByScore(front, func(p sim.Point) float64 {
		return float64(pl.available(p))
	}) {
		pl.overwhelm(target)
	}
	return pl.result()
}
