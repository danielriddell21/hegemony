package strategy

import (
	"math/rand/v2"

	"github.com/danielriddell21/hegemony/internal/sim"
)

type blob struct{}

func Blob() sim.Strategy { return blob{} }

func (blob) Name() string { return "Blob" }

func (blob) Move(v sim.View, _ *rand.Rand) []sim.Action {
	budget := v.Budget()
	frontier := v.Frontier()
	if budget <= 0 || len(frontier) == 0 {
		return nil
	}

	actions := make([]sim.Action, 0, len(frontier))
	for _, cell := range frontier {
		c := v.At(cell)
		cost := captureCost(c.Strength)
		if cost > budget {
			continue
		}
		actions = append(actions, sim.Action{
			Kind:   captureKind(c.Owner),
			Cell:   cell,
			Amount: cost,
		})
		budget -= cost
	}
	return actions
}
