package sim

import "slices"

// SnapshotConfig builds a world from an explicit board state rather than fresh
// spawns, so a strategy can fork the position it sees and roll it forward as a
// forward model. Policies assigns a strategy to each faction id present.
type SnapshotConfig struct {
	Width    int
	Height   int
	Cells    []Cell
	Params   Params
	Seed     uint64
	Policies map[FactionID]Strategy
}

func NewSnapshotWorld(cfg SnapshotConfig) *World {
	board := &Board{
		Width:  cfg.Width,
		Height: cfg.Height,
		cells:  append([]Cell(nil), cfg.Cells...),
	}

	maxID := Neutral
	ids := make([]FactionID, 0, len(cfg.Policies))
	for id := range cfg.Policies {
		ids = append(ids, id)
		maxID = max(maxID, id)
	}
	for _, c := range board.cells {
		maxID = max(maxID, c.Owner)
	}
	slices.Sort(ids)

	counts := make([]int, int(maxID)+1)
	for _, c := range board.cells {
		counts[c.Owner]++
	}

	w := &World{
		board:   board,
		params:  cfg.Params,
		counts:  counts,
		contest: newStream(cfg.Seed, contestStream),
		order:   newStream(cfg.Seed, orderStream),
	}
	for _, id := range ids {
		w.factions = append(w.factions, &faction{
			id:       id,
			strategy: cfg.Policies[id],
			rng:      newStream(cfg.Seed, uint64(id)),
		})
	}
	return w
}
