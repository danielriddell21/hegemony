package sim

import "testing"

func TestBoardInBounds(t *testing.T) {
	b := NewBoard(4, 3)
	cases := []struct {
		p    Point
		want bool
	}{
		{Point{0, 0}, true},
		{Point{3, 2}, true},
		{Point{4, 2}, false},
		{Point{3, 3}, false},
		{Point{-1, 0}, false},
	}
	for _, c := range cases {
		if got := b.InBounds(c.p); got != c.want {
			t.Errorf("InBounds(%v) = %v, want %v", c.p, got, c.want)
		}
	}
}

func TestBoardSetAtRoundTrip(t *testing.T) {
	b := NewBoard(5, 5)
	p := Point{2, 3}
	b.set(p, Cell{Owner: 2, Strength: 7})
	if got := b.At(p); got.Owner != 2 || got.Strength != 7 {
		t.Fatalf("At(%v) = %+v, want owner 2 strength 7", p, got)
	}
	if other := b.At(Point{0, 0}); other.Owner != Neutral {
		t.Fatalf("unrelated cell mutated: %+v", other)
	}
}

func TestBoardNeighbors(t *testing.T) {
	b := NewBoard(3, 3)
	if got := len(b.Neighbors(Point{0, 0})); got != 2 {
		t.Errorf("corner neighbors = %d, want 2", got)
	}
	if got := len(b.Neighbors(Point{1, 1})); got != 4 {
		t.Errorf("center neighbors = %d, want 4", got)
	}
	if got := len(b.Neighbors(Point{1, 0})); got != 3 {
		t.Errorf("edge neighbors = %d, want 3", got)
	}
}
