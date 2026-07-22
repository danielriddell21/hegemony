package sim

import (
	"slices"

	"github.com/danielriddell21/crucible/rng"
)

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
		contest: rng.Stream(cfg.Seed, contestStream),
		order:   rng.Stream(cfg.Seed, orderStream),
	}
	for _, id := range ids {
		w.factions = append(w.factions, &faction{
			id:       id,
			strategy: cfg.Policies[id],
			rng:      rng.Stream(cfg.Seed, uint64(id)),
		})
	}
	return w
}
