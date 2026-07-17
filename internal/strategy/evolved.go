package strategy

import (
	"math/rand/v2"

	"github.com/danielriddell21/hegemony/internal/sim"
)

type Weights struct {
	Weak      float64
	Open      float64
	Influence float64
	EnemyDist float64
	Source    float64
}

var evolvedWeights = Weights{
	Weak:      -1.563,
	Open:      2.744,
	Influence: 0.619,
	EnemyDist: -3.407,
	Source:    1.269,
}

type weighted struct {
	w Weights
}

func Evolved() sim.Strategy { return Weighted(evolvedWeights) }

func Weighted(w Weights) sim.Strategy { return weighted{w: w} }

func (weighted) Name() string { return "Evolved" }

func (s weighted) Move(v sim.View, _ *rand.Rand) []sim.Action {
	front := v.Frontier()
	if len(front) == 0 {
		return nil
	}
	_, enemies := scanOwners(v)
	return captureByScore(v, front, func(p sim.Point) float64 {
		f := featuresOf(v, p, enemies)
		return s.w.Weak*f.weak +
			s.w.Open*f.open +
			s.w.Influence*f.influence +
			s.w.EnemyDist*f.enemyDist +
			s.w.Source*f.source
	})
}

type features struct {
	weak      float64
	open      float64
	influence float64
	enemyDist float64
	source    float64
}

func featuresOf(v sim.View, p sim.Point, enemies []sim.Point) features {
	open := 0
	me := v.Faction()
	source := 0
	for _, q := range v.Neighbors(p) {
		c := v.At(q)
		if c.Owner == sim.Neutral {
			open++
		}
		if c.Owner == me && c.Strength > source {
			source = c.Strength
		}
	}
	enemyDist := 0.0
	if len(enemies) > 0 {
		enemyDist = float64(minDistance(p, enemies))
	}
	return features{
		weak:      -float64(v.At(p).Strength),
		open:      float64(open),
		influence: localInfluence(v, p),
		enemyDist: enemyDist,
		source:    float64(source),
	}
}
