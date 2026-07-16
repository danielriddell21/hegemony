package strategy_test

import (
	"math/rand/v2"
	"reflect"
	"testing"

	"github.com/danielriddell21/hegemony/internal/sim"
	"github.com/danielriddell21/hegemony/internal/strategy"
)

type fakeView struct {
	w, h     int
	faction  sim.FactionID
	budget   int
	cells    map[sim.Point]sim.Cell
	owned    []sim.Point
	frontier []sim.Point
}

func (v fakeView) Width() int             { return v.w }
func (v fakeView) Height() int            { return v.h }
func (v fakeView) Faction() sim.FactionID { return v.faction }
func (v fakeView) Budget() int            { return v.budget }
func (v fakeView) Owned() []sim.Point     { return v.owned }
func (v fakeView) Frontier() []sim.Point  { return v.frontier }

func (v fakeView) At(p sim.Point) sim.Cell {
	if c, ok := v.cells[p]; ok {
		return c
	}
	return sim.Cell{Owner: sim.Neutral, Strength: 3}
}

func (v fakeView) InBounds(p sim.Point) bool {
	return p.X >= 0 && p.X < v.w && p.Y >= 0 && p.Y < v.h
}

func (v fakeView) Neighbors(p sim.Point) []sim.Point {
	var out []sim.Point
	for _, o := range []sim.Point{{X: 0, Y: -1}, {X: 0, Y: 1}, {X: -1, Y: 0}, {X: 1, Y: 0}} {
		q := sim.Point{X: p.X + o.X, Y: p.Y + o.Y}
		if v.InBounds(q) {
			out = append(out, q)
		}
	}
	return out
}

func mixedFrontier() fakeView {
	cells := map[sim.Point]sim.Cell{
		{X: 1, Y: 0}: {Owner: sim.Neutral, Strength: 8},
		{X: 2, Y: 0}: {Owner: sim.Neutral, Strength: 2},
		{X: 3, Y: 0}: {Owner: 2, Strength: 5},
	}
	return fakeView{
		w: 8, h: 4, faction: 1, budget: 40, cells: cells,
		owned:    []sim.Point{{X: 0, Y: 0}},
		frontier: []sim.Point{{X: 1, Y: 0}, {X: 2, Y: 0}, {X: 3, Y: 0}},
	}
}

func allStrategies(t *testing.T) []sim.Strategy {
	t.Helper()
	names := strategy.Names()
	out := make([]sim.Strategy, 0, len(names))
	for _, name := range names {
		s, ok := strategy.New(name)
		if !ok {
			t.Fatalf("registry advertises %q but New rejects it", name)
		}
		out = append(out, s)
	}
	return out
}

func TestStrategiesProduceLegalActions(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	for _, s := range allStrategies(t) {
		v := mixedFrontier()
		front := frontierSet(v)
		owned := pointSet(v.owned)
		spent := 0
		for _, a := range s.Move(v, rng) {
			if a.Amount <= 0 {
				t.Errorf("%s: non-positive amount %d", s.Name(), a.Amount)
			}
			if a.Kind == sim.Reinforce {
				if _, ok := owned[a.Cell]; !ok {
					t.Errorf("%s: reinforce targets non-owned cell %v", s.Name(), a.Cell)
				}
			} else {
				if _, ok := front[a.Cell]; !ok {
					t.Errorf("%s: %v targets non-frontier cell %v", s.Name(), a.Kind, a.Cell)
				}
				if want := kindFor(v.At(a.Cell).Owner); a.Kind != want {
					t.Errorf("%s: kind = %v for owner %d, want %v", s.Name(), a.Kind, v.At(a.Cell).Owner, want)
				}
			}
			spent += a.Amount
		}
		if spent > v.budget {
			t.Errorf("%s: spent %d over budget %d", s.Name(), spent, v.budget)
		}
	}
}

func TestLookaheadPlaysAPanelPlan(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 1))
	v := mixedFrontier()
	got := strategy.Lookahead().Move(v, rng)
	if len(got) == 0 {
		t.Fatal("lookahead returned no actions on a non-empty frontier")
	}
	panel := [][]sim.Action{
		strategy.Greedy().Move(mixedFrontier(), rng),
		strategy.Blob().Move(mixedFrontier(), rng),
		strategy.Frontier().Move(mixedFrontier(), rng),
		strategy.Influence().Move(mixedFrontier(), rng),
		strategy.Voronoi().Move(mixedFrontier(), rng),
	}
	for _, plan := range panel {
		if reflect.DeepEqual(got, plan) {
			return
		}
	}
	t.Errorf("lookahead played a plan none of its panel produced: %+v", got)
}

func TestHeadhunterAdvancesOnWeakestFaction(t *testing.T) {
	// Faction 2 is smaller than faction 3, so headhunter should push toward
	// faction 2 — the frontier cell nearest its cell (3,1) goes first.
	v := fakeView{
		w: 8, h: 3, faction: 1, budget: 200,
		cells: map[sim.Point]sim.Cell{
			{X: 1, Y: 1}: {Owner: 1, Strength: 5},
			{X: 3, Y: 1}: {Owner: 2, Strength: 4},
			{X: 5, Y: 1}: {Owner: 3, Strength: 4},
			{X: 6, Y: 1}: {Owner: 3, Strength: 4},
			{X: 7, Y: 1}: {Owner: 3, Strength: 4},
		},
		owned:    []sim.Point{{X: 1, Y: 1}},
		frontier: []sim.Point{{X: 0, Y: 1}, {X: 2, Y: 1}, {X: 1, Y: 0}, {X: 1, Y: 2}},
	}
	actions := strategy.Headhunter().Move(v, rand.New(rand.NewPCG(1, 1)))
	if len(actions) == 0 {
		t.Fatal("headhunter returned no actions")
	}
	if actions[0].Cell != (sim.Point{X: 2, Y: 1}) {
		t.Fatalf("first target = %v, want (2,1) advancing toward faction 2", actions[0].Cell)
	}
}

func TestVoronoiClaimsAwayFromEnemy(t *testing.T) {
	// Budget affords one capture; voronoi must not grab the cell nearest the
	// enemy at (0,1).
	v := fakeView{
		w: 5, h: 3, faction: 1, budget: 4,
		cells: map[sim.Point]sim.Cell{
			{X: 2, Y: 1}: {Owner: 1, Strength: 5},
			{X: 0, Y: 1}: {Owner: 2, Strength: 5},
		},
		owned:    []sim.Point{{X: 2, Y: 1}},
		frontier: []sim.Point{{X: 1, Y: 1}, {X: 3, Y: 1}, {X: 2, Y: 0}, {X: 2, Y: 2}},
	}
	actions := strategy.Voronoi().Move(v, rand.New(rand.NewPCG(1, 1)))
	if len(actions) != 1 {
		t.Fatalf("got %d actions, want 1 within budget", len(actions))
	}
	if actions[0].Cell == (sim.Point{X: 1, Y: 1}) {
		t.Error("voronoi grabbed the border cell nearest the enemy first")
	}
}

func TestBulwarkFortifiesContestedBorder(t *testing.T) {
	// An owned cell adjacent to an enemy should draw a Reinforce; the quiet
	// owned cell should not.
	v := fakeView{
		w: 6, h: 3, faction: 1, budget: 60,
		cells: map[sim.Point]sim.Cell{
			{X: 2, Y: 1}: {Owner: 1, Strength: 5},
			{X: 4, Y: 1}: {Owner: 1, Strength: 5},
			{X: 5, Y: 1}: {Owner: 2, Strength: 9},
		},
		owned:    []sim.Point{{X: 2, Y: 1}, {X: 4, Y: 1}},
		frontier: []sim.Point{{X: 3, Y: 1}},
	}
	reinforced := make(map[sim.Point]bool)
	for _, a := range strategy.Bulwark().Move(v, rand.New(rand.NewPCG(1, 1))) {
		if a.Kind == sim.Reinforce {
			reinforced[a.Cell] = true
		}
	}
	if !reinforced[sim.Point{X: 4, Y: 1}] {
		t.Error("expected the enemy-adjacent cell (4,1) to be reinforced")
	}
	if reinforced[sim.Point{X: 2, Y: 1}] {
		t.Error("quiet cell (2,1) should not be reinforced")
	}
}

func TestGreedyTargetsWeakestFirst(t *testing.T) {
	v := mixedFrontier()
	actions := strategy.Greedy().Move(v, rand.New(rand.NewPCG(1, 1)))
	if len(actions) == 0 {
		t.Fatal("greedy returned no actions")
	}
	if actions[0].Cell != (sim.Point{X: 2, Y: 0}) {
		t.Fatalf("first target = %v, want weakest (2,0)", actions[0].Cell)
	}
}

func TestEmptyFrontierYieldsNoActions(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 1))
	v := fakeView{w: 4, h: 4, faction: 1, budget: 100, cells: map[sim.Point]sim.Cell{}}
	for _, s := range allStrategies(t) {
		if got := s.Move(v, rng); len(got) != 0 {
			t.Errorf("%s: got %d actions on empty frontier, want 0", s.Name(), len(got))
		}
	}
}

func TestRegistry(t *testing.T) {
	names := strategy.Names()
	if len(names) != 9 {
		t.Fatalf("Names() = %v, want 9 entries", names)
	}
	if _, ok := strategy.New("greedy"); !ok {
		t.Error("New is not case-insensitive for \"greedy\"")
	}
	if _, ok := strategy.New("headhunter"); !ok {
		t.Error("New does not resolve \"headhunter\"")
	}
	if _, ok := strategy.New("nope"); ok {
		t.Error("New(\"nope\") returned ok, want false")
	}
}

func frontierSet(v fakeView) map[sim.Point]struct{} {
	return pointSet(v.frontier)
}

func pointSet(pts []sim.Point) map[sim.Point]struct{} {
	out := make(map[sim.Point]struct{}, len(pts))
	for _, p := range pts {
		out[p] = struct{}{}
	}
	return out
}

func kindFor(owner sim.FactionID) sim.ActionKind {
	if owner == sim.Neutral {
		return sim.Expand
	}
	return sim.Attack
}
