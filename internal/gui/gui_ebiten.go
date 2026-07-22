//go:build ebiten

package gui

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/danielriddell21/crucible/record"

	"github.com/danielriddell21/hegemony/internal/sim"
	"github.com/danielriddell21/hegemony/internal/strategy"
)

const (
	hudLineHeight = 16
	shadeCeiling  = 40
	mapHeaderH    = 20
	boardWidth    = 240
	boardMaxRows  = 16
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
	ebiten.SetWindowSize(w, h)
	ebiten.SetWindowTitle("hegemony — war map")
	configureWindow(cfg)
	if err := ebiten.RunGame(g); err != nil {
		return fmt.Errorf("run war map: %w", err)
	}
	return nil
}

func runBoard(cfg Config) error {
	g := &boardGame{cfg: cfg, palette: palette(), link: cfg.Link}
	w, h := g.Layout(0, 0)
	ebiten.SetWindowSize(w, h)
	ebiten.SetWindowTitle("hegemony — leaderboard")
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
	pix      []byte
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
	screen.Fill(color.RGBA{R: 18, G: 18, B: 22, A: 255})
	board := g.world.Board()
	cs := float32(g.cfg.CellSize)
	for y := range board.Height {
		for x := range board.Width {
			c := board.At(sim.Point{X: x, Y: y})
			vector.FillRect(screen, float32(x)*cs, float32(y)*cs, cs-1, cs-1, colorFor(g.palette, c), false)
		}
	}
	ebitenutil.DebugPrintAt(screen, g.statusLine(), 4, g.cfg.Height*g.cfg.CellSize+4)

	if g.link == nil {
		// Standalone: no separate leaderboard window, so draw the standings here.
		for i, id := range g.world.Factions() {
			line := fmt.Sprintf("%-10s %5.1f%%", g.world.StrategyName(id), 100*g.world.Shares()[id])
			ebitenutil.DebugPrintAt(screen, line, 4, g.cfg.Height*g.cfg.CellSize+mapHeaderH+i*hudLineHeight)
		}
	}

	if g.rec != nil && !g.rec.Done() {
		b := screen.Bounds()
		if g.pix == nil {
			g.pix = make([]byte, 4*b.Dx()*b.Dy())
		}
		screen.ReadPixels(g.pix)
		g.rec.Add(&image.RGBA{Pix: g.pix, Stride: 4 * b.Dx(), Rect: image.Rect(0, 0, b.Dx(), b.Dy())})
	}
}

func (g *mapGame) statusLine() string {
	if !g.over {
		return fmt.Sprintf("tick %d   running", g.world.TickCount())
	}
	if g.winner != sim.Neutral {
		return fmt.Sprintf("tick %d   winner: %s", g.world.TickCount(), g.world.StrategyName(g.winner))
	}
	return fmt.Sprintf("tick %d   ended", g.world.TickCount())
}

func (g *mapGame) Layout(_, _ int) (int, int) {
	h := g.cfg.Height*g.cfg.CellSize + mapHeaderH
	if g.link == nil {
		h += len(g.cfg.Strategies) * hudLineHeight
	}
	return g.cfg.Width * g.cfg.CellSize, h
}

type boardGame struct {
	cfg     Config
	palette []color.RGBA
	link    *Link
	state   Msg
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
	screen.Fill(color.RGBA{R: 18, G: 18, B: 22, A: 255})
	ebitenutil.DebugPrintAt(screen, g.header(), 6, 6)

	stats := append([]FactionStat(nil), g.state.Factions...)
	sort.SliceStable(stats, func(i, j int) bool { return stats[i].Share > stats[j].Share })
	for i, s := range stats {
		y := 6 + (i+2)*hudLineHeight
		clr := g.palette[(s.ID-1)%len(g.palette)]
		vector.FillRect(screen, 6, float32(y)+1, 10, 10, clr, false)
		mark := " "
		if g.state.Over && s.ID == g.state.Winner {
			mark = "*"
		}
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%s %-10s %5.1f%%", mark, s.Name, 100*s.Share), 20, y)
	}
}

func (g *boardGame) header() string {
	if g.state.Over {
		return fmt.Sprintf("tick %d   ended", g.state.Tick)
	}
	return fmt.Sprintf("tick %d   running", g.state.Tick)
}

func (g *boardGame) Layout(_, _ int) (int, int) {
	return boardWidth, (boardMaxRows + 3) * hudLineHeight
}

func colorFor(pal []color.RGBA, c sim.Cell) color.RGBA {
	if c.Owner == sim.Neutral {
		return color.RGBA{R: 40, G: 40, B: 48, A: 255}
	}
	base := pal[(int(c.Owner)-1)%len(pal)]
	f := 0.45 + 0.55*clampF(float64(c.Strength)/shadeCeiling)
	return color.RGBA{
		R: uint8(float64(base.R) * f),
		G: uint8(float64(base.G) * f),
		B: uint8(float64(base.B) * f),
		A: 255,
	}
}

func sameState(a, b Msg) bool {
	return a.Tick == b.Tick && a.Over == b.Over && a.Winner == b.Winner
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
