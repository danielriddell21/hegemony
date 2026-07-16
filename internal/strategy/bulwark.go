package strategy

import (
	"math/rand/v2"

	"github.com/danielriddell21/hegemony/internal/sim"
)

const fortifyStep = 8

type bulwark struct{}

func Bulwark() sim.Strategy { return bulwark{} }

func (bulwark) Name() string { return "Bulwark" }

func (bulwark) Move(v sim.View, _ *rand.Rand) []sim.Action {
	front := v.Frontier()
	pl := newPlan(v)

	// Expand where it is cheap, weakest-first.
	for _, target := range sortedByScore(front, func(p sim.Point) float64 {
		return -float64(v.At(p).Strength)
	}) {
		pl.capture(target)
	}

	// Thicken only the cells that actually face an enemy, pulling strength in
	// from calmer owned neighbours so the contested front is costly to crack.
	me := v.Faction()
	contested := make([]sim.Point, 0)
	for _, p := range v.Owned() {
		if touchesEnemy(v, p, me) {
			contested = append(contested, p)
		}
	}
	for _, cell := range sortedByScore(contested, func(p sim.Point) float64 {
		return -float64(v.At(p).Strength)
	}) {
		pl.reinforce(cell, fortifyStep)
	}
	return pl.result()
}

func touchesEnemy(v sim.View, p sim.Point, me sim.FactionID) bool {
	for _, q := range v.Neighbors(p) {
		if o := v.At(q).Owner; o != sim.Neutral && o != me {
			return true
		}
	}
	return false
}
