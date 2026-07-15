package sim

type Params struct {
	IncomeBase      int
	IncomePerCell   int
	NeutralDefense  int
	StartStrength   int
	MaxCellStrength int
	Jitter          float64
}

func DefaultParams() Params {
	return Params{
		IncomeBase:      2,
		IncomePerCell:   1,
		NeutralDefense:  3,
		StartStrength:   20,
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
