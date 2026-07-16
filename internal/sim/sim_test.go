package sim

import (
	"math/rand/v2"
	"reflect"
	"testing"
)

type fixedStrategy struct {
	name string
	move func(v View) []Action
}

func (f fixedStrategy) Name() string { return f.name }

func (f fixedStrategy) Move(v View, _ *rand.Rand) []Action { return f.move(v) }

func noopStrategy(name string) Strategy {
	return fixedStrategy{name: name, move: func(View) []Action { return nil }}
}

func exactParams() Params {
	p := DefaultParams()
	p.Jitter = 0
	return p
}

func TestMoveCapturesNeutral(t *testing.T) {
	w := NewWorld(MatchConfig{
		Width: 6, Height: 6, Seed: 1, Params: exactParams(),
		Entrants: []Entrant{{Strategy: noopStrategy("A"), Spawn: Point{2, 2}}},
	})
	f := w.factions[0]
	from, to := Point{2, 2}, Point{3, 2}
	before := w.board.At(from).Strength
	w.apply(f, Action{From: from, To: to, Amount: 4})
	if got := w.board.At(to); got.Owner != f.id {
		t.Fatalf("target owner = %d, want %d", got.Owner, f.id)
	}
	if got := w.board.At(to).Strength; got != 1 {
		t.Fatalf("garrison = %d, want 1 (4-3)", got)
	}
	if got := w.board.At(from).Strength; got != before-4 {
		t.Fatalf("source strength = %d, want %d", got, before-4)
	}
	if w.counts[f.id] != 2 {
		t.Fatalf("territory = %d, want 2", w.counts[f.id])
	}
}

func TestMoveToOwnedReinforces(t *testing.T) {
	w := NewWorld(MatchConfig{
		Width: 4, Height: 4, Seed: 1, Params: exactParams(),
		Entrants: []Entrant{{Strategy: noopStrategy("A"), Spawn: Point{1, 1}}},
	})
	f := w.factions[0]
	w.setOwner(Point{1, 2}, f.id, 2)
	w.apply(f, Action{From: Point{1, 1}, To: Point{1, 2}, Amount: 5})
	if got := w.board.At(Point{1, 2}).Strength; got != 7 {
		t.Fatalf("reinforced strength = %d, want 7", got)
	}
}

func TestMoveRejectedWhenNotAdjacent(t *testing.T) {
	w := NewWorld(MatchConfig{
		Width: 6, Height: 6, Seed: 1, Params: exactParams(),
		Entrants: []Entrant{{Strategy: noopStrategy("A"), Spawn: Point{2, 2}}},
	})
	f := w.factions[0]
	before := w.board.At(Point{2, 2}).Strength
	w.apply(f, Action{From: Point{2, 2}, To: Point{5, 5}, Amount: 5})
	if w.counts[f.id] != 1 {
		t.Fatalf("territory = %d, want 1 (non-adjacent rejected)", w.counts[f.id])
	}
	if got := w.board.At(Point{2, 2}).Strength; got != before {
		t.Fatalf("source strength = %d, want unchanged %d", got, before)
	}
}

func TestMoveRejectedWhenOverSourceStrength(t *testing.T) {
	w := NewWorld(MatchConfig{
		Width: 4, Height: 4, Seed: 1, Params: exactParams(),
		Entrants: []Entrant{{Strategy: noopStrategy("A"), Spawn: Point{1, 1}}},
	})
	f := w.factions[0]
	w.setStrength(Point{1, 1}, 3)
	w.apply(f, Action{From: Point{1, 1}, To: Point{1, 2}, Amount: 10})
	if got := w.board.At(Point{1, 1}).Strength; got != 3 {
		t.Fatalf("source strength = %d, want unchanged 3", got)
	}
	if w.board.At(Point{1, 2}).Owner != Neutral {
		t.Fatal("target captured despite insufficient source strength")
	}
}

func TestReinforceCapsStrength(t *testing.T) {
	p := exactParams()
	p.MaxCellStrength = 25
	w := NewWorld(MatchConfig{
		Width: 4, Height: 4, Seed: 1, Params: p,
		Entrants: []Entrant{{Strategy: noopStrategy("A"), Spawn: Point{1, 1}}},
	})
	f := w.factions[0]
	w.setStrength(Point{1, 1}, 40)
	w.setOwner(Point{1, 2}, f.id, 24)
	w.apply(f, Action{From: Point{1, 1}, To: Point{1, 2}, Amount: 10})
	if got := w.board.At(Point{1, 2}).Strength; got != 25 {
		t.Fatalf("strength = %d, want capped at 25", got)
	}
}

func TestIncomeGrowsOwnedCells(t *testing.T) {
	w := NewWorld(MatchConfig{
		Width: 5, Height: 5, Seed: 1, Params: exactParams(),
		Entrants: []Entrant{{Strategy: noopStrategy("A"), Spawn: Point{2, 2}}},
	})
	before := w.board.At(Point{2, 2}).Strength
	w.applyIncome()
	if got := w.board.At(Point{2, 2}).Strength; got != before+1 {
		t.Fatalf("owned strength after income = %d, want %d", got, before+1)
	}
	if got := w.board.At(Point{0, 0}).Strength; got != exactParams().NeutralDefense {
		t.Fatalf("neutral cell grew to %d, want %d", got, exactParams().NeutralDefense)
	}
}

func TestResolveContestDeterministicAndExact(t *testing.T) {
	w := NewWorld(MatchConfig{
		Width: 2, Height: 2, Seed: 1, Params: exactParams(),
		Entrants: []Entrant{{Strategy: noopStrategy("A"), Spawn: Point{0, 0}}},
	})
	if win, rem := w.resolveContest(10, 4); !win || rem != 6 {
		t.Fatalf("contest(10,4) jitter 0 = (%v,%d), want (true,6)", win, rem)
	}
	if win, rem := w.resolveContest(3, 4); win || rem != 1 {
		t.Fatalf("contest(3,4) jitter 0 = (%v,%d), want (false,1)", win, rem)
	}
}

func TestLastFactionStandingWins(t *testing.T) {
	res := RunMatch(MatchConfig{
		Width: 5, Height: 5, Seed: 7, MaxTicks: 500, WinThreshold: 0.6,
		Params: exactParams(),
		Entrants: []Entrant{
			{Strategy: sweeper("A"), Spawn: Point{0, 0}},
			{Strategy: noopStrategy("B"), Spawn: Point{4, 4}},
		},
	})
	if res.Winner != 1 {
		t.Fatalf("winner = %d, want faction 1", res.Winner)
	}
}

func TestRunMatchReproducible(t *testing.T) {
	cfg := MatchConfig{
		Width: 8, Height: 8, Seed: 42, MaxTicks: 200, WinThreshold: 0.6,
		Params: DefaultParams(),
		Entrants: []Entrant{
			{Strategy: sweeper("A"), Spawn: Point{1, 1}},
			{Strategy: sweeper("B"), Spawn: Point{6, 6}},
		},
	}
	a := RunMatch(cfg)
	b := RunMatch(cfg)
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("same seed produced different results:\n%+v\n%+v", a, b)
	}
}

func sweeper(name string) Strategy {
	return fixedStrategy{name: name, move: func(v View) []Action {
		out := make([]Action, 0)
		for _, p := range v.Owned() {
			strength := v.At(p).Strength
			for _, q := range v.Neighbors(p) {
				if strength <= 5 {
					break
				}
				if v.At(q).Owner == v.Faction() {
					continue
				}
				out = append(out, Action{From: p, To: q, Amount: 5})
				strength -= 5
			}
		}
		return out
	}}
}

func TestSnapshotWorldForksIndependently(t *testing.T) {
	src := NewWorld(MatchConfig{
		Width: 6, Height: 6, Seed: 1, Params: exactParams(),
		Entrants: []Entrant{
			{Strategy: noopStrategy("A"), Spawn: Point{1, 1}},
			{Strategy: noopStrategy("B"), Spawn: Point{4, 4}},
		},
	})
	cells := make([]Cell, 6*6)
	for y := range 6 {
		for x := range 6 {
			cells[y*6+x] = src.board.At(Point{x, y})
		}
	}
	snap := NewSnapshotWorld(SnapshotConfig{
		Width: 6, Height: 6, Cells: cells, Params: exactParams(), Seed: 1,
		Policies: map[FactionID]Strategy{1: noopStrategy("A"), 2: noopStrategy("B")},
	})

	if snap.Territory(1) != src.Territory(1) || snap.Territory(2) != src.Territory(2) {
		t.Fatalf("snapshot territory %d/%d, want %d/%d",
			snap.Territory(1), snap.Territory(2), src.Territory(1), src.Territory(2))
	}
	if snap.board.At(Point{1, 1}).Strength != src.board.At(Point{1, 1}).Strength {
		t.Fatal("snapshot did not preserve cell strength")
	}
	snap.setStrength(Point{1, 1}, 99)
	if src.board.At(Point{1, 1}).Strength == 99 {
		t.Fatal("snapshot aliased the source board")
	}
}

func TestSpreadSpawnsDistinctInBounds(t *testing.T) {
	pts := SpreadSpawns(10, 10, 6)
	if len(pts) != 6 {
		t.Fatalf("got %d spawns, want 6", len(pts))
	}
	seen := make(map[Point]struct{})
	for _, p := range pts {
		if p.X < 0 || p.X >= 10 || p.Y < 0 || p.Y >= 10 {
			t.Errorf("spawn %v out of bounds", p)
		}
		if _, dup := seen[p]; dup {
			t.Errorf("duplicate spawn %v", p)
		}
		seen[p] = struct{}{}
	}
}
