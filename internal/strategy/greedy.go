package strategy

import (
	"cmp"
	"math/rand/v2"
	"slices"

	"github.com/danielriddell21/hegemony/internal/sim"
)

type greedy struct{}

func Greedy() sim.Strategy { return greedy{} }

func (greedy) Name() string { return "Greedy" }

func (greedy) Move(v sim.View, _ *rand.Rand) []sim.Action {
	budget := v.Budget()
	frontier := slices.Clone(v.Frontier())
	if budget <= 0 || len(frontier) == 0 {
		return nil
	}

	slices.SortFunc(frontier, func(a, b sim.Point) int {
		if c := cmp.Compare(v.At(a).Strength, v.At(b).Strength); c != 0 {
			return c
		}
		if c := cmp.Compare(a.Y, b.Y); c != 0 {
			return c
		}
		return cmp.Compare(a.X, b.X)
	})

	actions := make([]sim.Action, 0, len(frontier))
	for _, cell := range frontier {
		c := v.At(cell)
		need := captureCost(c.Strength)
		if need > budget {
			break
		}
		actions = append(actions, sim.Action{
			Kind:   captureKind(c.Owner),
			Cell:   cell,
			Amount: need,
		})
		budget -= need
	}
	return actions
}

func captureCost(strength int) int {
	return strength + strength/4 + 1
}
