package sim

import "math/rand/v2"

type faction struct {
	id       FactionID
	strategy Strategy
	rng      *rand.Rand
}

type World struct {
	board    *Board
	params   Params
	factions []*faction
	counts   []int
	contest  *rand.Rand
	tick     int
}

func NewWorld(cfg MatchConfig) *World {
	board := NewBoard(cfg.Width, cfg.Height)
	for y := range board.Height {
		for x := range board.Width {
			board.set(Point{X: x, Y: y}, Cell{Owner: Neutral, Strength: cfg.Params.NeutralDefense})
		}
	}

	counts := make([]int, len(cfg.Entrants)+1)
	counts[Neutral] = board.Width * board.Height

	w := &World{
		board:   board,
		params:  cfg.Params,
		counts:  counts,
		contest: newStream(cfg.Seed, contestStream),
	}
	for i, e := range cfg.Entrants {
		id := FactionID(i + 1)
		w.factions = append(w.factions, &faction{
			id:       id,
			strategy: e.Strategy,
			rng:      newStream(cfg.Seed, uint64(id)),
		})
		w.setOwner(e.Spawn, id, cfg.Params.StartStrength)
	}
	return w
}

func (w *World) Tick() {
	w.applyIncome()
	for _, f := range w.factions {
		if w.counts[f.id] == 0 {
			continue
		}
		v := w.viewFor(f)
		for _, a := range f.strategy.Move(v, f.rng) {
			w.apply(f, a)
		}
	}
	w.tick++
}

func (w *World) applyIncome() {
	for y := range w.board.Height {
		for x := range w.board.Width {
			p := Point{X: x, Y: y}
			if c := w.board.At(p); c.Owner != Neutral {
				w.setStrength(p, c.Strength+w.params.IncomePerCell)
			}
		}
	}
}

func (w *World) viewFor(f *faction) *boardView {
	owned := make([]Point, 0, w.counts[f.id])
	for y := range w.board.Height {
		for x := range w.board.Width {
			p := Point{X: x, Y: y}
			if w.board.At(p).Owner == f.id {
				owned = append(owned, p)
			}
		}
	}
	return &boardView{
		board:    w.board,
		faction:  f.id,
		owned:    owned,
		frontier: w.frontierOf(f.id, owned),
	}
}

func (w *World) frontierOf(id FactionID, owned []Point) []Point {
	seen := make(map[Point]struct{})
	frontier := make([]Point, 0)
	for _, p := range owned {
		for _, q := range w.board.Neighbors(p) {
			if w.board.At(q).Owner == id {
				continue
			}
			if _, ok := seen[q]; ok {
				continue
			}
			seen[q] = struct{}{}
			frontier = append(frontier, q)
		}
	}
	return frontier
}

func (w *World) apply(f *faction, a Action) {
	if a.Amount <= 0 || !w.board.InBounds(a.From) || !w.board.InBounds(a.To) || !adjacent(a.From, a.To) {
		return
	}
	src := w.board.At(a.From)
	if src.Owner != f.id || a.Amount > src.Strength {
		return
	}

	w.setStrength(a.From, src.Strength-a.Amount)
	dst := w.board.At(a.To)
	if dst.Owner == f.id {
		w.setStrength(a.To, dst.Strength+a.Amount)
		return
	}
	if win, remaining := w.resolveContest(a.Amount, dst.Strength); win {
		w.setOwner(a.To, f.id, w.capped(remaining))
	} else {
		w.setStrength(a.To, remaining)
	}
}

func adjacent(a, b Point) bool {
	dx, dy := a.X-b.X, a.Y-b.Y
	return dx*dx+dy*dy == 1
}

func (w *World) resolveContest(atk, def int) (bool, int) {
	noise := 1 + w.params.Jitter*(w.contest.Float64()*2-1)
	eff := int(float64(atk)*noise + 0.5)
	if eff > def {
		return true, max(1, eff-def)
	}
	return false, max(1, def-eff)
}

func (w *World) setOwner(p Point, owner FactionID, strength int) {
	prev := w.board.At(p).Owner
	if prev != owner {
		w.counts[prev]--
		w.counts[owner]++
	}
	w.board.set(p, Cell{Owner: owner, Strength: strength})
}

func (w *World) setStrength(p Point, strength int) {
	c := w.board.At(p)
	c.Strength = w.capped(strength)
	w.board.set(p, c)
}

func (w *World) capped(strength int) int {
	if w.params.MaxCellStrength > 0 && strength > w.params.MaxCellStrength {
		return w.params.MaxCellStrength
	}
	return strength
}
