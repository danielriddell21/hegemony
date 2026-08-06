//go:build ebiten

package gui

import (
	"errors"
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/danielriddell21/crucible/canvas"
	"github.com/danielriddell21/crucible/record"
	"github.com/danielriddell21/crucible/window"

	"github.com/danielriddell21/hegemony/internal/sim"
	"github.com/danielriddell21/hegemony/internal/strategy"
)

func Available() bool { return true }

func Run(cfg Config) error {
	if cfg.Role == RoleBoard {
		return runBoard(cfg)
	}
	return runMap(cfg)
}

func configureWindow(cfg Config) {
	if cfg.OffsetIndex > 0 {
		ebiten.SetWindowPosition(80+cfg.OffsetIndex*40, 80+cfg.OffsetIndex*40)
	}
	if cfg.Link != nil {
		// Coordinated windows must keep updating while unfocused so a background
		// window stays in sync and notices the leader closing (its Update must run
		// to drain the closed link and terminate).
		ebiten.SetRunnableOnUnfocused(true)
	}
}

func runMap(cfg Config) error {
	world, err := buildWorld(cfg)
	if err != nil {
		return err
	}
	g := &mapGame{cfg: cfg, world: world, palette: palette(), link: cfg.Link}
	if cfg.Rec.Recording() {
		g.rec = record.New(cfg.Rec)
	}
	if g.link != nil {
		g.lastSent = g.shared() // suppress an initial publish; the child starts empty and fills in
	}
	w, h := g.Layout(0, 0)
	window.Configure(window.Options{Title: "hegemony — war map", Width: w, Height: h, MinWidth: w / 2, MinHeight: h / 2})
	configureWindow(cfg)
	if err := ebiten.RunGame(g); err != nil {
		return fmt.Errorf("run war map: %w", err)
	}
	return nil
}

func runBoard(cfg Config) error {
	g := &boardGame{cfg: cfg, palette: palette(), link: cfg.Link}
	w, h := g.Layout(0, 0)
	window.Configure(window.Options{Title: "hegemony — leaderboard", Width: w, Height: h, MinWidth: w / 2, MinHeight: h / 2})
	configureWindow(cfg)
	if err := ebiten.RunGame(g); err != nil {
		return fmt.Errorf("run leaderboard: %w", err)
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

type mapGame struct {
	cfg      Config
	world    *sim.World
	palette  []color.RGBA
	acc      float64
	over     bool
	winner   sim.FactionID
	link     *Link
	lastSent Msg
	rec      *record.Recorder
	canvas   *canvas.Canvas
}

func (g *mapGame) Update() error {
	if g.rec != nil && g.rec.Done() {
		if err := g.rec.Save(g.cfg.Rec.Path); err != nil {
			return fmt.Errorf("save recording: %w", err)
		}
		return ebiten.Termination
	}
	if !g.over {
		g.acc += float64(g.cfg.TicksPerSecond) / float64(ebiten.TPS())
		for g.acc >= 1 && !g.over {
			g.acc--
			g.step()
		}
	}
	return g.syncLink()
}

func (g *mapGame) step() {
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

func (g *mapGame) syncLink() error {
	if g.link == nil {
		return nil
	}
	for done := false; !done; {
		select {
		case m, ok := <-g.link.In:
			if !ok || m.Type == "quit" {
				return ebiten.Termination
			}
		default:
			done = true
		}
	}
	if cur := g.shared(); !sameState(cur, g.lastSent) {
		if g.send(cur) {
			g.lastSent = cur
		}
	}
	return nil
}

func (g *mapGame) shared() Msg {
	ids := g.world.Factions()
	shares := g.world.Shares()
	stats := make([]FactionStat, 0, len(ids))
	for _, id := range ids {
		stats = append(stats, FactionStat{ID: int(id), Name: g.world.StrategyName(id), Share: shares[id]})
	}
	return Msg{Type: "state", Tick: g.world.TickCount(), Over: g.over, Winner: int(g.winner), Factions: stats}
}

func (g *mapGame) send(m Msg) bool {
	if g.link == nil || g.link.Out == nil {
		return false
	}
	select {
	case g.link.Out <- m:
		return true
	default:
		return false
	}
}

func (g *mapGame) Draw(screen *ebiten.Image) {
	if g.canvas == nil {
		w, h := MapSize(g.cfg, g.link == nil)
		g.canvas = canvas.New(w, h)
	}
	DrawMap(g.canvas, g.cfg, MapView{
		World:     g.world,
		Over:      g.over,
		Winner:    g.winner,
		Standings: g.link == nil, // no separate leaderboard window
	}, g.palette)
	screen.WritePixels(g.canvas.Pixels())

	if g.rec != nil && !g.rec.Done() {
		w, h := g.canvas.Size()
		g.rec.Add(record.FromRGBA(g.canvas.Pixels(), w, h))
	}
}

func (g *mapGame) Layout(_, _ int) (int, int) {
	return MapSize(g.cfg, g.link == nil)
}

type boardGame struct {
	cfg     Config
	palette []color.RGBA
	link    *Link
	state   Msg
	canvas  *canvas.Canvas
}

func (g *boardGame) Update() error {
	if g.link == nil {
		return nil
	}
	for done := false; !done; {
		select {
		case m, ok := <-g.link.In:
			if !ok || m.Type == "quit" {
				return ebiten.Termination
			}
			if m.Type == "state" {
				g.state = m
			}
		default:
			done = true
		}
	}
	return nil
}

func (g *boardGame) Draw(screen *ebiten.Image) {
	if g.canvas == nil {
		g.canvas = canvas.New(g.Layout(0, 0))
	}
	DrawBoard(g.canvas, g.state, g.palette)
	screen.WritePixels(g.canvas.Pixels())
}

func (g *boardGame) Layout(_, _ int) (int, int) {
	return boardWidth, (boardMaxRows + 3) * hudLineHeight
}

func sameState(a, b Msg) bool {
	return a.Tick == b.Tick && a.Over == b.Over && a.Winner == b.Winner
}
