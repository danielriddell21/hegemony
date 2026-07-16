package strategy

import (
	"math/rand/v2"

	"github.com/danielriddell21/hegemony/internal/sim"
)

const attackPriority = 1e6

type headhunter struct{}

func Headhunter() sim.Strategy { return headhunter{} }

func (headhunter) Name() string { return "Headhunter" }

func (headhunter) Move(v sim.View, _ *rand.Rand) []sim.Action {
	front := v.Frontier()
	if v.Budget() <= 0 || len(front) == 0 {
		return nil
	}

	counts, enemies := scanOwners(v)
	if len(enemies) == 0 {
		return captureByScore(v, front, func(p sim.Point) float64 { return roi(v, p) })
	}

	target := weakestEnemy(counts, v.Faction())
	targetCells := cellsOwnedBy(v, enemies, target)

	// Drive toward the smallest surviving faction and eat it: eliminating a
	// faction shrinks the field and pushes the last-standing win condition.
	return captureByScore(v, front, func(p sim.Point) float64 {
		if v.At(p).Owner == target {
			return attackPriority
		}
		return -float64(minDistance(p, targetCells))
	})
}
