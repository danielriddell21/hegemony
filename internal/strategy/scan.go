package strategy

import "github.com/danielriddell21/hegemony/internal/sim"

func scanOwners(v sim.View) (counts map[sim.FactionID]int, enemyCells []sim.Point) {
	me := v.Faction()
	counts = make(map[sim.FactionID]int)
	enemyCells = make([]sim.Point, 0)
	for y := range v.Height() {
		for x := range v.Width() {
			p := sim.Point{X: x, Y: y}
			owner := v.At(p).Owner
			counts[owner]++
			if owner != sim.Neutral && owner != me {
				enemyCells = append(enemyCells, p)
			}
		}
	}
	return counts, enemyCells
}

func cellsOwnedBy(v sim.View, cells []sim.Point, owner sim.FactionID) []sim.Point {
	out := make([]sim.Point, 0)
	for _, p := range cells {
		if v.At(p).Owner == owner {
			out = append(out, p)
		}
	}
	return out
}

func weakestEnemy(counts map[sim.FactionID]int, me sim.FactionID) sim.FactionID {
	best := sim.Neutral
	bestCount := 0
	for id, n := range counts {
		if id == sim.Neutral || id == me || n == 0 {
			continue
		}
		if best == sim.Neutral || n < bestCount || (n == bestCount && id < best) {
			best, bestCount = id, n
		}
	}
	return best
}

func onFrontier(v sim.View, p sim.Point, me sim.FactionID) bool {
	for _, q := range v.Neighbors(p) {
		if v.At(q).Owner != me {
			return true
		}
	}
	return false
}

func manhattan(a, b sim.Point) int {
	return abs(a.X-b.X) + abs(a.Y-b.Y)
}

func minDistance(p sim.Point, targets []sim.Point) int {
	best := -1
	for _, t := range targets {
		if d := manhattan(p, t); best < 0 || d < best {
			best = d
		}
	}
	return best
}
