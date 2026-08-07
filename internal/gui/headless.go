package gui

import (
	"errors"
	"fmt"
	"image"

	"github.com/danielriddell21/crucible/canvas"
	"github.com/danielriddell21/crucible/demo"
	"github.com/danielriddell21/crucible/record"

	"github.com/danielriddell21/hegemony/internal/sim"
	"github.com/danielriddell21/hegemony/internal/strategy"
)

// framesPerSecond is the window's update rate, which the headless recorder
// matches so both advance the match at the same pace.
const framesPerSecond = 60

// Render records a match to cfg.Rec.Path without opening a window. The war map
// is composed by the same [DrawMap] the window uses, so the media is identical
// to what a viewer sees — and it needs no display.
//
// The file extension picks the format: .gif or .mp4.
func Render(cfg Config) error {
	world, err := newWorld(cfg)
	if err != nil {
		return err
	}
	w, h := MapSize(cfg, true)
	c := canvas.New(w, h)
	pal := palette()
	// The map only changes at its frontiers, and every colour it uses is known
	// up front: a scene palette plus delta frames keeps the GIF small where a
	// dithered full-frame encode would run to tens of megabytes.
	rec := record.New(cfg.Rec,
		record.WithPalette(demo.Ramp(append(pal, background, colText, neutralCell), shadeSteps)),
		record.WithFrameDiff(),
	)

	view := MapView{World: world, Standings: true}
	// Advance exactly as the window does: a fraction of a tick per frame, so a
	// recording runs the match at the same pace whichever path drew it.
	acc := 0.0
	clip := demo.Clip{
		Frames: cfg.Rec.Frames,
		Step: func(int) error {
			if view.Over {
				return nil // hold on the finished map for the remaining frames
			}
			for acc += float64(cfg.TicksPerSecond) / framesPerSecond; acc >= 1 && !view.Over; acc-- {
				switch id, done := world.Decided(cfg.WinThreshold); {
				case done:
					view.Over, view.Winner = true, id
				case world.TickCount() >= cfg.MaxTicks:
					view.Over = true
				default:
					world.Tick()
				}
			}
			return nil
		},
		Frame: func(int) image.Image {
			DrawMap(c, cfg, view, pal)
			return record.FromRGBA(c.Pixels(), w, h)
		},
	}
	if _, err := clip.Record(rec); err != nil {
		return fmt.Errorf("capture match: %w", err)
	}
	if err := rec.Save(cfg.Rec.Path); err != nil {
		return fmt.Errorf("save recording: %w", err)
	}
	fmt.Printf("%s: %d frames\n", cfg.Rec.Path, rec.Len())
	return nil
}

// newWorld builds the match world from the configured strategies. It mirrors
// the windowed build so both paths simulate the same match.
func newWorld(cfg Config) (*sim.World, error) {
	names := cfg.Strategies
	if len(names) == 0 {
		names = strategy.Names()
	}
	if len(names) == 0 {
		return nil, errors.New("no strategies to run")
	}
	spawns := sim.SpreadSpawns(cfg.Width, cfg.Height, len(names))
	entrants := make([]sim.Entrant, 0, len(names))
	for i, name := range names {
		s, ok := strategy.New(name)
		if !ok {
			return nil, fmt.Errorf("unknown strategy %q", name)
		}
		entrants = append(entrants, sim.Entrant{Strategy: s, Spawn: spawns[i]})
	}
	return sim.NewWorld(sim.MatchConfig{
		Width:        cfg.Width,
		Height:       cfg.Height,
		Seed:         cfg.Seed,
		MaxTicks:     cfg.MaxTicks,
		WinThreshold: cfg.WinThreshold,
		Params:       cfg.Params,
		Entrants:     entrants,
	}), nil
}
