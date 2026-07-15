package sim

import "math/rand/v2"

type Strategy interface {
	Name() string
	Move(v View, rng *rand.Rand) []Action
}
