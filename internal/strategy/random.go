package strategy

import (
	"math/rand/v2"

	"github.com/danielriddell21/hegemony/internal/sim"
)

type random struct{}

func Random() sim.Strategy { return random{} }

func (random) Name() string { return "Random" }

func (random) Move(v sim.View, rng *rand.Rand) []sim.Action {
	budget := v.Budget()
	frontier := v.Frontier()
	if budget <= 0 || len(frontier) == 0 {
		return nil
	}

	actions := make([]sim.Action, 0, len(frontier))
	for _, i := range rng.Perm(len(frontier)) {
		if budget <= 0 {
			break
		}
		cell := frontier[i]
		amount := 1 + rng.IntN(budget)
		actions = append(actions, sim.Action{
			Kind:   captureKind(v.At(cell).Owner),
			Cell:   cell,
			Amount: amount,
		})
		budget -= amount
	}
	return actions
}

func captureKind(owner sim.FactionID) sim.ActionKind {
	if owner == sim.Neutral {
		return sim.Expand
	}
	return sim.Attack
}
