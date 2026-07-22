package sim

// contestStream and orderStream keep the two shared simulation streams
// decorrelated from each other and from the per-member streams keyed by id.
const (
	contestStream uint64 = 0xC047E57
	orderStream   uint64 = 0x0DE40DE4
)
