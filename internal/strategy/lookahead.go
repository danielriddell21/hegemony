package strategy

import (
	"math/rand/v2"

	"github.com/danielriddell21/hegemony/internal/sim"
)

const (
	potentialWeight = 0.15
	exposureWeight  = 0.05
)

type lookahead struct {
	panel []sim.Strategy
}

func Lookahead() sim.Strategy {
	return lookahead{panel: []sim.Strategy{Greedy(), Blob(), Frontier(), Influence(), Voronoi()}}
}

func (lookahead) Name() string { return "Lookahead" }

func (l lookahead) Move(v sim.View, rng *rand.Rand) []sim.Action {
	if v.Budget() <= 0 || len(v.Frontier()) == 0 {
		return nil
	}

	var best []sim.Action
	var bestScore float64
	found := false
	for _, s := range l.panel {
		candidate := s.Move(v, rng)
		if len(candidate) == 0 {
			continue
		}
		if score := evaluate(v, candidate); !found || score > bestScore {
			found, bestScore, best = true, score, candidate
		}
	}
	return best
}

func evaluate(v sim.View, actions []sim.Action) float64 {
	sb := newSandbox(v)
	for _, a := range actions {
		sb.apply(a)
	}
	return sb.score()
}

type sandbox struct {
	w, h  int
	me    sim.FactionID
	owner []sim.FactionID
	str   []int
}

func newSandbox(v sim.View) *sandbox {
	w, h := v.Width(), v.Height()
	sb := &sandbox{
		w: w, h: h, me: v.Faction(),
		owner: make([]sim.FactionID, w*h),
		str:   make([]int, w*h),
	}
	for y := range h {
		for x := range w {
			c := v.At(sim.Point{X: x, Y: y})
			i := y*w + x
			sb.owner[i] = c.Owner
			sb.str[i] = c.Strength
		}
	}
	return sb
}

func (sb *sandbox) apply(a sim.Action) {
	i := a.Cell.Y*sb.w + a.Cell.X
	if a.Kind == sim.Reinforce {
		sb.str[i] += a.Amount
		return
	}
	if a.Amount > sb.str[i] {
		sb.owner[i] = sb.me
		sb.str[i] = max(1, a.Amount-sb.str[i])
		return
	}
	sb.str[i] -= a.Amount
}

func (sb *sandbox) score() float64 {
	var territory, potential, exposure int
	for i := range sb.owner {
		if sb.owner[i] != sb.me {
			continue
		}
		territory++
		p, e := sb.borderOf(i)
		potential += p
		exposure += e
	}
	return float64(territory) + potentialWeight*float64(potential) - exposureWeight*float64(exposure)
}

func (sb *sandbox) borderOf(i int) (potential, exposure int) {
	x, y := i%sb.w, i/sb.w
	for _, n := range sb.neighbors(x, y) {
		switch o := sb.owner[n]; {
		case o == sim.Neutral:
			potential++
		case o != sb.me:
			exposure += sb.str[n]
		}
	}
	return potential, exposure
}

func (sb *sandbox) neighbors(x, y int) []int {
	out := make([]int, 0, 4)
	if x > 0 {
		out = append(out, y*sb.w+x-1)
	}
	if x < sb.w-1 {
		out = append(out, y*sb.w+x+1)
	}
	if y > 0 {
		out = append(out, (y-1)*sb.w+x)
	}
	if y < sb.h-1 {
		out = append(out, (y+1)*sb.w+x)
	}
	return out
}
