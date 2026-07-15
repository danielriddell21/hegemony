//go:build ebiten

package gui

import (
	"errors"
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/danielriddell21/hegemony/internal/sim"
	"github.com/danielriddell21/hegemony/internal/strategy"
)

const (
	hudLineHeight = 16
	shadeCeiling  = 40
)

func Available() bool { return true }

func Run(cfg Config) error {
	world, err := buildWorld(cfg)
	if err != nil {
		return err
	}
	g := &game{cfg: cfg, world: world, palette: palette()}
	w, h := g.Layout(0, 0)
	ebiten.SetWindowSize(w, h)
	ebiten.SetWindowTitle("hegemony")
	if err := ebiten.RunGame(g); err != nil {
		return fmt.Errorf("ebiten: %w", err)
	}
	return nil
}

func buildWorld(cfg Config) (*sim.World, error) {
	if len(cfg.Strategies) == 0 {
		return nil, errors.New("no strategies to run")
	}
	spawns := sim.SpreadSpawns(cfg.Width, cfg.Height, len(cfg.Strategies))
	entrants := make([]sim.Entrant, 0, len(cfg.Strategies))
	for i, name := range cfg.Strategies {
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

type game struct {
	cfg     Config
	world   *sim.World
	palette []color.RGBA
	acc     float64
	over    bool
	winner  sim.FactionID
}

func (g *game) Update() error {
	if g.over {
		return nil
	}
	g.acc += float64(g.cfg.TicksPerSecond) / float64(ebiten.TPS())
	for g.acc >= 1 && !g.over {
		g.acc--
		g.step()
	}
	return nil
}

func (g *game) step() {
	if id, done := g.world.Decided(g.cfg.WinThreshold); done {
		g.over, g.winner = true, id
		return
	}
	if g.world.TickCount() >= g.cfg.MaxTicks {
		g.over = true
		return
	}
	g.world.Tick()
}

func (g *game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 18, G: 18, B: 22, A: 255})
	board := g.world.Board()
	cs := float32(g.cfg.CellSize)
	for y := range board.Height {
		for x := range board.Width {
			c := board.At(sim.Point{X: x, Y: y})
			vector.FillRect(screen, float32(x)*cs, float32(y)*cs, cs-1, cs-1, g.colorFor(c), false)
		}
	}
	g.drawHUD(screen)
}

func (g *game) drawHUD(screen *ebiten.Image) {
	y := g.cfg.Height*g.cfg.CellSize + 4
	status := "running"
	if g.over {
		status = "ended"
		if g.winner != sim.Neutral {
			status = fmt.Sprintf("winner: %s", g.world.StrategyName(g.winner))
		}
	}
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("tick %d   %s", g.world.TickCount(), status), 4, y)
	for i, id := range g.world.Factions() {
		line := fmt.Sprintf("%-10s %5.1f%%", g.world.StrategyName(id), 100*g.world.Shares()[id])
		ebitenutil.DebugPrintAt(screen, line, 4, y+(i+1)*hudLineHeight)
	}
}

func (g *game) Layout(_, _ int) (int, int) {
	width := g.cfg.Width * g.cfg.CellSize
	height := g.cfg.Height*g.cfg.CellSize + (len(g.cfg.Strategies)+2)*hudLineHeight
	return width, height
}

func (g *game) colorFor(c sim.Cell) color.RGBA {
	if c.Owner == sim.Neutral {
		return color.RGBA{R: 40, G: 40, B: 48, A: 255}
	}
	base := g.palette[(int(c.Owner)-1)%len(g.palette)]
	f := 0.45 + 0.55*clampF(float64(c.Strength)/shadeCeiling)
	return color.RGBA{
		R: uint8(float64(base.R) * f),
		G: uint8(float64(base.G) * f),
		B: uint8(float64(base.B) * f),
		A: 255,
	}
}

func clampF(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func palette() []color.RGBA {
	return []color.RGBA{
		{R: 232, G: 93, B: 117, A: 255},
		{R: 86, G: 156, B: 214, A: 255},
		{R: 152, G: 195, B: 121, A: 255},
		{R: 229, G: 192, B: 123, A: 255},
		{R: 198, G: 120, B: 221, A: 255},
		{R: 86, G: 182, B: 194, A: 255},
	}
}
