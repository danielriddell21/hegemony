package strategy

import (
	"math/rand/v2"

	"github.com/danielriddell21/hegemony/internal/sim"
)

const influenceRadius = 3

type influence struct{}

func Influence() sim.Strategy { return influence{} }

func (influence) Name() string { return "Influence" }

func (influence) Move(v sim.View, _ *rand.Rand) []sim.Action {
	front := v.Frontier()
	if v.Budget() <= 0 || len(front) == 0 {
		return nil
	}
	return captureByScore(v, front, func(p sim.Point) float64 { return localInfluence(v, p) })
}

func localInfluence(v sim.View, p sim.Point) float64 {
	me := v.Faction()
	var sum float64
	for dy := -influenceRadius; dy <= influenceRadius; dy++ {
		for dx := -influenceRadius; dx <= influenceRadius; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			q := sim.Point{X: p.X + dx, Y: p.Y + dy}
			if !v.InBounds(q) {
				continue
			}
			owner := v.At(q).Owner
			if owner == sim.Neutral {
				continue
			}
			weight := float64(v.At(q).Strength) / float64(abs(dx)+abs(dy))
			if owner == me {
				sum += weight
			} else {
				sum -= weight
			}
		}
	}
	return sum
}
