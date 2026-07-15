package sim

type FactionID int

const Neutral FactionID = 0

type Point struct {
	X, Y int
}

type Cell struct {
	Owner    FactionID
	Strength int
}

type Board struct {
	Width  int
	Height int
	cells  []Cell
}

func NewBoard(width, height int) *Board {
	return &Board{
		Width:  width,
		Height: height,
		cells:  make([]Cell, width*height),
	}
}

func (b *Board) index(p Point) int { return p.Y*b.Width + p.X }

func (b *Board) InBounds(p Point) bool {
	return p.X >= 0 && p.X < b.Width && p.Y >= 0 && p.Y < b.Height
}

func (b *Board) At(p Point) Cell { return b.cells[b.index(p)] }

func (b *Board) set(p Point, c Cell) { b.cells[b.index(p)] = c }

var neighborOffsets = [4]Point{{X: 0, Y: -1}, {X: 0, Y: 1}, {X: -1, Y: 0}, {X: 1, Y: 0}}

func (b *Board) Neighbors(p Point) []Point {
	out := make([]Point, 0, 4)
	for _, o := range neighborOffsets {
		q := Point{X: p.X + o.X, Y: p.Y + o.Y}
		if b.InBounds(q) {
			out = append(out, q)
		}
	}
	return out
}
