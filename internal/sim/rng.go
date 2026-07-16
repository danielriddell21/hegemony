package sim

import "math/rand/v2"

const (
	contestStream uint64 = 0xC047E57
	orderStream   uint64 = 0x0DE40DE4
)

func newStream(seed, stream uint64) *rand.Rand {
	return rand.New(rand.NewPCG(seed, stream))
}
