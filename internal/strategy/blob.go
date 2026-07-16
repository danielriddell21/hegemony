package strategy

import (
	"math/rand/v2"

	"github.com/danielriddell21/hegemony/internal/sim"
)

type blob struct{}

func Blob() sim.Strategy { return blob{} }

func (blob) Name() string { return "Blob" }

func (blob) Move(v sim.View, _ *rand.Rand) []sim.Action {
	front := v.Frontier()
	if len(front) == 0 {
		return nil
	}
	// Uniform score means captures fall in a spatial sweep, so the mass grows
	// outward on every front at once.
	return captureByScore(v, front, func(sim.Point) float64 { return 0 })
}
