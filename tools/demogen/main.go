// Command demogen renders hegemony's documentation media headlessly: short,
// deterministic war-map clips. It drives the simulation and composes frames on
// a software canvas — the same code the window draws with — so it needs no
// display.
//
// Regenerate every asset under docs/demos with:
//
//	just demos      // or: go run ./tools/demogen
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/danielriddell21/crucible/record"

	"github.com/danielriddell21/hegemony/internal/gui"
	"github.com/danielriddell21/hegemony/internal/sim"
)

const outDir = "docs/demos"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "demogen:", err)
		os.Exit(1)
	}
}

func run() error {
	if err := os.MkdirAll(outDir, 0o750); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}
	for _, c := range clips() {
		if err := c.record(); err != nil {
			return fmt.Errorf("%s: %w", c.name, err)
		}
	}
	return nil
}

// clip is one recorded match: who enters, on what board, for how long.
type clip struct {
	name string
	// ext is the output extension; an .mp4 records video instead of a GIF.
	ext        string
	seed       uint64
	width      int
	height     int
	strategies []string
	frames     int
}

// clips is the documentation set: a full free-for-all, and a three-way duel
// that shows contrasting doctrines head to head.
func clips() []clip {
	return []clip{
		{name: "match", ext: ".gif", seed: 3, width: 32, height: 32, frames: 200},
		{
			name: "duel", ext: ".gif", seed: 7, width: 32, height: 32, frames: 160,
			strategies: []string{"Blitzkrieg", "Turtle", "Bulwark"},
		},
	}
}

func (c clip) record() error {
	path := filepath.Join(outDir, c.name+c.ext)
	cfg := gui.Config{
		Width:          c.width,
		Height:         c.height,
		Seed:           c.seed,
		MaxTicks:       2000,
		WinThreshold:   0.6,
		Params:         sim.DefaultParams(),
		Strategies:     c.strategies,
		CellSize:       18,
		TicksPerSecond: 12,
		Rec:            record.Options{Path: path, Frames: c.frames, FPS: 30, Scale: 1},
	}
	if err := gui.Render(cfg); err != nil {
		return fmt.Errorf("render: %w", err)
	}
	return nil
}
