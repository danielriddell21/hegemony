package sim

type ActionKind uint8

const (
	Expand ActionKind = iota
	Attack
	Reinforce
)

func (k ActionKind) String() string {
	switch k {
	case Expand:
		return "expand"
	case Attack:
		return "attack"
	case Reinforce:
		return "reinforce"
	default:
		return "unknown"
	}
}

type Action struct {
	Kind   ActionKind
	Cell   Point
	Amount int
}
