package sim

type Params struct {
	IncomePerCell   int
	NeutralDefense  int
	StartStrength   int
	MaxCellStrength int
	Jitter          float64
}

func DefaultParams() Params {
	return Params{
		IncomePerCell:   1,
		NeutralDefense:  3,
		StartStrength:   25,
		MaxCellStrength: 99,
		Jitter:          0.25,
	}
}

type Entrant struct {
	Strategy Strategy
	Spawn    Point
}

type MatchConfig struct {
	Width        int
	Height       int
	Seed         uint64
	MaxTicks     int
	WinThreshold float64
	Params       Params
	Entrants     []Entrant
}
