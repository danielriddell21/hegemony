package strategy

import (
	"math/rand/v2"

	"github.com/danielriddell21/hegemony/internal/sim"
)

const (
	turtleFortifyStep   = 20
	turtleExpandCeiling = 3
)

type turtle struct{}

func Turtle() sim.Strategy { return turtle{} }

func (turtle) Name() string { return "Turtle" }

func (turtle) Move(v sim.View, _ *rand.Rand) []sim.Action {
	pl := newPlan(v)
	me := v.Faction()

	// Take only the cheapest neutral ground — never pick a fight.
	for _, target := range sortedByScore(v.Frontier(), func(p sim.Point) float64 {
		return -float64(v.At(p).Strength)
	}) {
		if c := v.At(target); c.Owner == sim.Neutral && c.Strength <= turtleExpandCeiling {
			pl.capture(target)
		}
	}

	// Pour everything else into the perimeter, building walls the enemy has to
	// bleed through.
	for _, p := range v.Owned() {
		if onFrontier(v, p, me) {
			pl.reinforce(p, turtleFortifyStep)
		}
	}
	return pl.result()
}
