package sim

type View interface {
	Width() int
	Height() int
	Faction() FactionID
	At(p Point) Cell
	InBounds(p Point) bool
	Neighbors(p Point) []Point
	Owned() []Point
	Frontier() []Point
}

type boardView struct {
	board    *Board
	faction  FactionID
	owned    []Point
	frontier []Point
}

func (v *boardView) Width() int                { return v.board.Width }
func (v *boardView) Height() int               { return v.board.Height }
func (v *boardView) Faction() FactionID        { return v.faction }
func (v *boardView) At(p Point) Cell           { return v.board.At(p) }
func (v *boardView) InBounds(p Point) bool     { return v.board.InBounds(p) }
func (v *boardView) Neighbors(p Point) []Point { return v.board.Neighbors(p) }
func (v *boardView) Owned() []Point            { return v.owned }
func (v *boardView) Frontier() []Point         { return v.frontier }
