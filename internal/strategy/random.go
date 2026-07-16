package strategy

import (
	"math/rand/v2"

	"github.com/danielriddell21/hegemony/internal/sim"
)

type random struct{}

func Random() sim.Strategy { return random{} }

func (random) Name() string { return "Random" }

func (random) Move(v sim.View, rng *rand.Rand) []sim.Action {
	front := v.Frontier()
	if len(front) == 0 {
		return nil
	}

	pl := newPlan(v)
	for _, i := range rng.Perm(len(front)) {
		target := front[i]
		avail := pl.available(target)
		if avail <= 0 {
			continue
		}
		pl.pressure(target, 1+rng.IntN(avail))
	}
	return pl.result()
}
