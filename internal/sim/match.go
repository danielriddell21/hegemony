package sim

type Result struct {
	Ticks     int
	Winner    FactionID
	Territory []int
	Names     []string
	Share     [][]float64
}

func RunMatch(cfg MatchConfig) Result {
	w := NewWorld(cfg)
	names := make([]string, len(cfg.Entrants)+1)
	for i, e := range cfg.Entrants {
		names[i+1] = e.Strategy.Name()
	}

	var share [][]float64
	for {
		share = append(share, w.Shares())
		if id, done := w.decided(cfg.WinThreshold); done {
			return w.finish(id, names, share)
		}
		if w.tick >= cfg.MaxTicks {
			return w.finish(w.leader(), names, share)
		}
		w.Tick()
	}
}

func (w *World) decided(threshold float64) (FactionID, bool) {
	total := w.board.Width * w.board.Height
	alive := 0
	var last FactionID
	for _, f := range w.factions {
		t := w.counts[f.id]
		if t == 0 {
			continue
		}
		alive++
		last = f.id
		if threshold > 0 && float64(t) >= threshold*float64(total) {
			return f.id, true
		}
	}
	if alive <= 1 {
		return last, true
	}
	return Neutral, false
}

func (w *World) leader() FactionID {
	best := Neutral
	bestCount := 0
	tie := false
	for _, f := range w.factions {
		switch t := w.counts[f.id]; {
		case t > bestCount:
			best, bestCount, tie = f.id, t, false
		case t == bestCount && t > 0:
			tie = true
		}
	}
	if tie {
		return Neutral
	}
	return best
}

func (w *World) finish(winner FactionID, names []string, share [][]float64) Result {
	terr := make([]int, len(w.counts))
	copy(terr, w.counts)
	return Result{
		Ticks:     w.tick,
		Winner:    winner,
		Territory: terr,
		Names:     names,
		Share:     share,
	}
}

func (w *World) Board() *Board { return w.board }

func (w *World) TickCount() int { return w.tick }

func (w *World) Territory(id FactionID) int { return w.counts[id] }

func (w *World) Shares() []float64 {
	total := float64(w.board.Width * w.board.Height)
	out := make([]float64, len(w.counts))
	for i, c := range w.counts {
		out[i] = float64(c) / total
	}
	return out
}

func (w *World) Factions() []FactionID {
	out := make([]FactionID, len(w.factions))
	for i, f := range w.factions {
		out[i] = f.id
	}
	return out
}

func (w *World) StrategyName(id FactionID) string {
	for _, f := range w.factions {
		if f.id == id {
			return f.strategy.Name()
		}
	}
	return ""
}

func (w *World) Decided(threshold float64) (FactionID, bool) {
	return w.decided(threshold)
}
