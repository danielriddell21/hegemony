package strategy

import (
	"cmp"
	"math/rand/v2"
	"slices"

	"github.com/danielriddell21/hegemony/internal/sim"
)

const fortifyStep = 6

type bulwark struct{}

func Bulwark() sim.Strategy { return bulwark{} }

func (bulwark) Name() string { return "Bulwark" }

func (bulwark) Move(v sim.View, _ *rand.Rand) []sim.Action {
	budget := v.Budget()
	if budget <= 0 {
		return nil
	}
	actions, spent := bulwarkExpand(v, budget*2/3)
	return bulwarkFortify(v, budget-spent, actions)
}

func bulwarkExpand(v sim.View, budget int) ([]sim.Action, int) {
	front := slices.Clone(v.Frontier())
	slices.SortFunc(front, func(a, b sim.Point) int {
		return cmp.Compare(v.At(a).Strength, v.At(b).Strength)
	})

	actions := make([]sim.Action, 0, len(front))
	spent := 0
	for _, cell := range front {
		c := v.At(cell)
		cost := captureCost(c.Strength)
		if cost > budget {
			break
		}
		actions = append(actions, sim.Action{Kind: captureKind(c.Owner), Cell: cell, Amount: cost})
		budget -= cost
		spent += cost
	}
	return actions, spent
}

func bulwarkFortify(v sim.View, budget int, actions []sim.Action) []sim.Action {
	me := v.Faction()
	border := make([]sim.Point, 0)
	for _, p := range v.Owned() {
		if touchesEnemy(v, p, me) {
			border = append(border, p)
		}
	}
	slices.SortFunc(border, func(a, b sim.Point) int {
		return cmp.Compare(v.At(a).Strength, v.At(b).Strength)
	})

	for _, cell := range border {
		if budget <= 0 {
			break
		}
		step := min(budget, fortifyStep)
		actions = append(actions, sim.Action{Kind: sim.Reinforce, Cell: cell, Amount: step})
		budget -= step
	}
	return actions
}

func touchesEnemy(v sim.View, p sim.Point, me sim.FactionID) bool {
	for _, q := range v.Neighbors(p) {
		if o := v.At(q).Owner; o != sim.Neutral && o != me {
			return true
		}
	}
	return false
}
