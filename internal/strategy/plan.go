package strategy

import "github.com/danielriddell21/hegemony/internal/sim"

type plan struct {
	v       sim.View
	avail   map[sim.Point]int
	actions []sim.Action
}

func newPlan(v sim.View) *plan {
	avail := make(map[sim.Point]int, len(v.Owned()))
	for _, p := range v.Owned() {
		avail[p] = v.At(p).Strength
	}
	return &plan{v: v, avail: avail, actions: make([]sim.Action, 0)}
}

func (pl *plan) result() []sim.Action {
	if len(pl.actions) == 0 {
		return nil
	}
	return pl.actions
}

func (pl *plan) capture(target sim.Point) {
	cost := captureCost(pl.v.At(target).Strength)
	src, best := pl.strongestSource(target)
	if best < cost {
		return
	}
	pl.spend(src, target, cost)
}

func (pl *plan) overwhelm(target sim.Point) {
	cost := captureCost(pl.v.At(target).Strength)
	src, best := pl.strongestSource(target)
	if best < cost {
		return
	}
	pl.spend(src, target, best)
}

func (pl *plan) pressure(target sim.Point, amount int) {
	src, best := pl.strongestSource(target)
	if best <= 0 {
		return
	}
	pl.spend(src, target, min(amount, best))
}

func (pl *plan) reinforce(target sim.Point, amount int) {
	src, best := pl.strongestSource(target)
	if best <= 0 {
		return
	}
	pl.spend(src, target, min(amount, best))
}

func (pl *plan) available(target sim.Point) int {
	_, best := pl.strongestSource(target)
	return best
}

func (pl *plan) spend(src, target sim.Point, amount int) {
	pl.avail[src] -= amount
	pl.actions = append(pl.actions, sim.Action{From: src, To: target, Amount: amount})
}

func (pl *plan) strongestSource(target sim.Point) (sim.Point, int) {
	var bestP sim.Point
	best := 0
	for _, q := range pl.v.Neighbors(target) {
		if pl.v.At(q).Owner != pl.v.Faction() {
			continue
		}
		if a := pl.avail[q]; a > best {
			best, bestP = a, q
		}
	}
	return bestP, best
}
