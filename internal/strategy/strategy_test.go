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
	cells    map[sim.Point]sim.Cell
	owned    []sim.Point
	frontier []sim.Point
}

func (v fakeView) Width() int             { return v.w }
func (v fakeView) Height() int            { return v.h }
func (v fakeView) Faction() sim.FactionID { return v.faction }
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

// mixedFrontier: faction 1 holds two strong cells with a mix of neutral and
// enemy cells on the frontier, each reachable from an owned neighbour.
func mixedFrontier() fakeView {
	return fakeView{
		w: 8, h: 5, faction: 1,
		cells: map[sim.Point]sim.Cell{
			{X: 1, Y: 2}: {Owner: 1, Strength: 40},
			{X: 2, Y: 2}: {Owner: 1, Strength: 40},
			{X: 3, Y: 2}: {Owner: 2, Strength: 6},
		},
		owned: []sim.Point{{X: 1, Y: 2}, {X: 2, Y: 2}},
		frontier: []sim.Point{
			{X: 0, Y: 2},
			{X: 1, Y: 1},
			{X: 1, Y: 3},
			{X: 2, Y: 1},
			{X: 2, Y: 3},
			{X: 3, Y: 2},
		},
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

func TestStrategiesProduceLegalMoves(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	for _, s := range allStrategies(t) {
		v := mixedFrontier()
		owned := pointSet(v.owned)
		front := pointSet(v.frontier)
		spent := make(map[sim.Point]int)
		for _, a := range s.Move(v, rng) {
			if a.Amount <= 0 {
				t.Errorf("%s: non-positive amount %d", s.Name(), a.Amount)
			}
			if _, ok := owned[a.From]; !ok {
				t.Errorf("%s: move from non-owned cell %v", s.Name(), a.From)
			}
			if manhattan(a.From, a.To) != 1 {
				t.Errorf("%s: move %v->%v is not to an adjacent cell", s.Name(), a.From, a.To)
			}
			_, toOwned := owned[a.To]
			_, toFront := front[a.To]
			if !toOwned && !toFront {
				t.Errorf("%s: move targets %v which is neither owned nor frontier", s.Name(), a.To)
			}
			spent[a.From] += a.Amount
		}
		for src, amt := range spent {
			if cap := v.At(src).Strength; amt > cap {
				t.Errorf("%s: moved %d out of %v holding only %d", s.Name(), amt, src, cap)
			}
		}
	}
}

func TestGreedyTargetsWeakestFirst(t *testing.T) {
	v := mixedFrontier()
	actions := strategy.Greedy().Move(v, rand.New(rand.NewPCG(1, 1)))
	if len(actions) == 0 {
		t.Fatal("greedy returned no actions")
	}
	if got := v.At(actions[0].To).Strength; got != 3 {
		t.Fatalf("first target strength = %d, want the weakest (3)", got)
	}
}

func TestEmptyFrontierYieldsNoActions(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 1))
	v := fakeView{w: 4, h: 4, faction: 1, cells: map[sim.Point]sim.Cell{}}
	for _, s := range allStrategies(t) {
		if got := s.Move(v, rng); len(got) != 0 {
			t.Errorf("%s: got %d actions on empty frontier, want 0", s.Name(), len(got))
		}
	}
}

func TestLookaheadPlaysAPanelPlan(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 1))
	got := strategy.Lookahead().Move(mixedFrontier(), rng)
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
	// Faction 2 (one cell) is weaker than faction 3 (three cells), so the
	// frontier cell nearest faction 2 at (4,1) goes first.
	v := fakeView{
		w: 9, h: 3, faction: 1,
		cells: map[sim.Point]sim.Cell{
			{X: 1, Y: 1}: {Owner: 1, Strength: 40},
			{X: 4, Y: 1}: {Owner: 2, Strength: 4},
			{X: 6, Y: 1}: {Owner: 3, Strength: 4},
			{X: 7, Y: 1}: {Owner: 3, Strength: 4},
			{X: 8, Y: 1}: {Owner: 3, Strength: 4},
		},
		owned:    []sim.Point{{X: 1, Y: 1}},
		frontier: []sim.Point{{X: 0, Y: 1}, {X: 2, Y: 1}, {X: 1, Y: 0}, {X: 1, Y: 2}},
	}
	actions := strategy.Headhunter().Move(v, rand.New(rand.NewPCG(1, 1)))
	if len(actions) == 0 {
		t.Fatal("headhunter returned no actions")
	}
	if actions[0].To != (sim.Point{X: 2, Y: 1}) {
		t.Fatalf("first target = %v, want (2,1) advancing toward faction 2", actions[0].To)
	}
}

func TestVoronoiClaimsAwayFromEnemy(t *testing.T) {
	// The source can afford exactly one capture; voronoi must not spend it on
	// the cell nearest the enemy at (0,1).
	v := fakeView{
		w: 5, h: 3, faction: 1,
		cells: map[sim.Point]sim.Cell{
			{X: 2, Y: 1}: {Owner: 1, Strength: 4},
			{X: 0, Y: 1}: {Owner: 2, Strength: 5},
		},
		owned:    []sim.Point{{X: 2, Y: 1}},
		frontier: []sim.Point{{X: 1, Y: 1}, {X: 3, Y: 1}, {X: 2, Y: 0}, {X: 2, Y: 2}},
	}
	actions := strategy.Voronoi().Move(v, rand.New(rand.NewPCG(1, 1)))
	if len(actions) != 1 {
		t.Fatalf("got %d actions, want 1 within source strength", len(actions))
	}
	if actions[0].To == (sim.Point{X: 1, Y: 1}) {
		t.Error("voronoi grabbed the border cell nearest the enemy first")
	}
}

func TestBulwarkFortifiesContestedBorder(t *testing.T) {
	// (3,1) is an interior cell; (4,1) touches the enemy at (5,1). Bulwark
	// should shift strength into (4,1).
	v := fakeView{
		w: 7, h: 3, faction: 1,
		cells: map[sim.Point]sim.Cell{
			{X: 3, Y: 1}: {Owner: 1, Strength: 40},
			{X: 4, Y: 1}: {Owner: 1, Strength: 6},
			{X: 5, Y: 1}: {Owner: 2, Strength: 9},
		},
		owned:    []sim.Point{{X: 3, Y: 1}, {X: 4, Y: 1}},
		frontier: []sim.Point{{X: 5, Y: 1}, {X: 4, Y: 0}, {X: 4, Y: 2}, {X: 3, Y: 0}, {X: 3, Y: 2}, {X: 2, Y: 1}},
	}
	fortified := false
	for _, a := range strategy.Bulwark().Move(v, rand.New(rand.NewPCG(1, 1))) {
		if a.To == (sim.Point{X: 4, Y: 1}) {
			fortified = true
		}
	}
	if !fortified {
		t.Error("expected bulwark to reinforce the enemy-adjacent cell (4,1)")
	}
}

func TestTurtleWallsItsPerimeter(t *testing.T) {
	v := mixedFrontier()
	owned := pointSet(v.owned)
	reinforced := false
	for _, a := range strategy.Turtle().Move(v, rand.New(rand.NewPCG(1, 1))) {
		if _, ok := owned[a.To]; ok {
			reinforced = true
		}
	}
	if !reinforced {
		t.Error("turtle never reinforced one of its own cells")
	}
}

func TestBlitzkriegConcentratesForce(t *testing.T) {
	v := mixedFrontier()
	actions := strategy.Blitzkrieg().Move(v, rand.New(rand.NewPCG(1, 1)))
	if len(actions) == 0 {
		t.Fatal("blitzkrieg returned no actions")
	}
	// A spearhead pours far more than the minimal capture cost of a weak cell.
	heaviest := 0
	for _, a := range actions {
		if a.Amount > heaviest {
			heaviest = a.Amount
		}
	}
	if heaviest <= 10 {
		t.Fatalf("heaviest strike = %d, want a concentrated push (>10)", heaviest)
	}
}

func TestRegistry(t *testing.T) {
	names := strategy.Names()
	if len(names) != 12 {
		t.Fatalf("Names() = %v, want 12 entries", names)
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

func pointSet(pts []sim.Point) map[sim.Point]struct{} {
	out := make(map[sim.Point]struct{}, len(pts))
	for _, p := range pts {
		out[p] = struct{}{}
	}
	return out
}

func manhattan(a, b sim.Point) int {
	dx := a.X - b.X
	if dx < 0 {
		dx = -dx
	}
	dy := a.Y - b.Y
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}
