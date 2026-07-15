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

func TestExpandCapturesNeutral(t *testing.T) {
	target := Point{3, 2}
	// Build a world directly for fine-grained control.
	w := NewWorld(MatchConfig{
		Width: 6, Height: 6, Seed: 1, Params: exactParams(),
		Entrants: []Entrant{{Strategy: noopStrategy("A"), Spawn: Point{2, 2}}},
	})
	f := w.factions[0]
	f.budget = 10
	w.apply(f, Action{Kind: Expand, Cell: target, Amount: 4})
	if got := w.board.At(target); got.Owner != f.id {
		t.Fatalf("target owner = %d, want %d", got.Owner, f.id)
	}
	if got := w.board.At(target).Strength; got != 1 {
		t.Fatalf("garrison = %d, want 1 (4-3)", got)
	}
	if f.budget != 6 {
		t.Fatalf("budget = %d, want 6", f.budget)
	}
	if w.counts[f.id] != 2 {
		t.Fatalf("territory = %d, want 2", w.counts[f.id])
	}
}

func TestExpandNonAdjacentRejected(t *testing.T) {
	w := NewWorld(MatchConfig{
		Width: 6, Height: 6, Seed: 1, Params: exactParams(),
		Entrants: []Entrant{{Strategy: noopStrategy("A"), Spawn: Point{2, 2}}},
	})
	f := w.factions[0]
	f.budget = 100
	w.apply(f, Action{Kind: Expand, Cell: Point{5, 5}, Amount: 50})
	if w.counts[f.id] != 1 {
		t.Fatalf("territory = %d, want 1 (non-adjacent rejected)", w.counts[f.id])
	}
	if f.budget != 100 {
		t.Fatalf("budget = %d, want 100 (no spend on rejected action)", f.budget)
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
	f.budget = 100
	w.apply(f, Action{Kind: Reinforce, Cell: Point{1, 1}, Amount: 50})
	if got := w.board.At(Point{1, 1}).Strength; got != 25 {
		t.Fatalf("strength = %d, want capped at 25", got)
	}
}

func TestAttackCapturesEnemy(t *testing.T) {
	w := NewWorld(MatchConfig{
		Width: 6, Height: 1, Seed: 1, Params: exactParams(),
		Entrants: []Entrant{
			{Strategy: noopStrategy("A"), Spawn: Point{2, 0}},
			{Strategy: noopStrategy("B"), Spawn: Point{3, 0}},
		},
	})
	a := w.factions[0]
	a.budget = 100
	// B's spawn has StartStrength 20; attack it with 21 (jitter 0 -> captures).
	w.apply(a, Action{Kind: Attack, Cell: Point{3, 0}, Amount: 21})
	if got := w.board.At(Point{3, 0}).Owner; got != a.id {
		t.Fatalf("captured owner = %d, want %d", got, a.id)
	}
	if w.counts[2] != 0 {
		t.Fatalf("faction B territory = %d, want 0", w.counts[2])
	}
}

func TestResolveContestDeterministicAndExact(t *testing.T) {
	w := NewWorld(MatchConfig{
		Width: 2, Height: 2, Seed: 1, Params: exactParams(),
		Entrants: []Entrant{{Strategy: noopStrategy("A"), Spawn: Point{0, 0}}},
	})
	win, rem := w.resolveContest(10, 4)
	if !win || rem != 6 {
		t.Fatalf("contest(10,4) jitter 0 = (%v,%d), want (true,6)", win, rem)
	}
	win, rem = w.resolveContest(3, 4)
	if win || rem != 1 {
		t.Fatalf("contest(3,4) jitter 0 = (%v,%d), want (false,1)", win, rem)
	}
}

func TestIncomeGrowsBudget(t *testing.T) {
	w := NewWorld(MatchConfig{
		Width: 5, Height: 5, Seed: 1, Params: exactParams(),
		Entrants: []Entrant{{Strategy: noopStrategy("A"), Spawn: Point{2, 2}}},
	})
	w.applyIncome()
	// base 2 + perCell 1 * 1 owned cell = 3.
	if got := w.factions[0].budget; got != 3 {
		t.Fatalf("budget after income = %d, want 3", got)
	}
}

func TestLastFactionStandingWins(t *testing.T) {
	// Faction A expands aggressively; B does nothing and holds one weak cell.
	grabRight := fixedStrategy{name: "A", move: func(v View) []Action {
		out := make([]Action, 0, len(v.Frontier()))
		budget := v.Budget()
		for _, cell := range v.Frontier() {
			c := v.At(cell)
			cost := c.Strength + 1
			if cost > budget {
				continue
			}
			kind := Expand
			if c.Owner != Neutral {
				kind = Attack
			}
			out = append(out, Action{Kind: kind, Cell: cell, Amount: cost})
			budget -= cost
		}
		return out
	}}
	res := RunMatch(MatchConfig{
		Width: 5, Height: 5, Seed: 7, MaxTicks: 500, WinThreshold: 0.6,
		Params: exactParams(),
		Entrants: []Entrant{
			{Strategy: grabRight, Spawn: Point{0, 0}},
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
			{Strategy: randomish("A"), Spawn: Point{1, 1}},
			{Strategy: randomish("B"), Spawn: Point{6, 6}},
		},
	}
	a := RunMatch(cfg)
	b := RunMatch(cfg)
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("same seed produced different results:\n%+v\n%+v", a, b)
	}
}

func randomish(name string) Strategy {
	return fixedStrategy{name: name, move: func(v View) []Action {
		front := v.Frontier()
		if len(front) == 0 || v.Budget() <= 0 {
			return nil
		}
		cell := front[0]
		c := v.At(cell)
		kind := Expand
		if c.Owner != Neutral {
			kind = Attack
		}
		return []Action{{Kind: kind, Cell: cell, Amount: v.Budget()}}
	}}
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
