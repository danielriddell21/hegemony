package strategy

import "github.com/danielriddell21/hegemony/internal/sim"

// plan tracks the strength still available to move out of each owned cell this
// tick and accumulates the resulting From→To moves.
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

// capture launches a decisive strike at target from its strongest adjacent
// owned cell, but only when that cell alone can pay the full capture cost.
func (pl *plan) capture(target sim.Point) {
	cost := captureCost(pl.v.At(target).Strength)
	src, best := pl.strongestSource(target)
	if best < cost {
		return
	}
	pl.spend(src, target, cost)
}

// overwhelm captures target and pours the whole source cell behind it, trading
// tempo for a durable garrison.
func (pl *plan) overwhelm(target sim.Point) {
	cost := captureCost(pl.v.At(target).Strength)
	src, best := pl.strongestSource(target)
	if best < cost {
		return
	}
	pl.spend(src, target, best)
}

// pressure moves up to amount strength at target from its strongest adjacent
// owned cell, even if that is not enough to capture — it still wears defenders
// down.
func (pl *plan) pressure(target sim.Point, amount int) {
	src, best := pl.strongestSource(target)
	if best <= 0 {
		return
	}
	pl.spend(src, target, min(amount, best))
}

// reinforce shifts up to amount strength into an owned cell from its strongest
// owned neighbour, thickening a chosen part of the line.
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
