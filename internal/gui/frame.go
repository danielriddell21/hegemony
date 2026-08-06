package gui

import (
	"fmt"
	"image/color"
	"sort"

	"github.com/danielriddell21/crucible/canvas"

	"github.com/danielriddell21/hegemony/internal/sim"
)

// textAscent lifts a top-left text origin onto the canvas face's baseline.
const textAscent = 11

// Frame metrics shared by the window and the headless recorder.
const (
	hudLineHeight = 16
	shadeCeiling  = 40
	mapHeaderH    = 20
	boardWidth    = 240
	boardMaxRows  = 16
)

// background is the colour behind every frame.
var background = color.RGBA{R: 18, G: 18, B: 22, A: 255}

// MapView is everything the war map draws: the world, how the match ended, and
// whether the standings belong on this window.
type MapView struct {
	World  *sim.World
	Over   bool
	Winner sim.FactionID
	// Standings draws the leaderboard inline, for a run with no separate
	// leaderboard window.
	Standings bool
}

// DrawMap composes the war map onto c. Drawing through the software canvas
// rather than the display means the same code paints a live window and a
// headless recording, pixel for pixel.
func DrawMap(c *canvas.Canvas, cfg Config, v MapView, pal []color.RGBA) {
	c.Fill(background)
	board := v.World.Board()
	cs := cfg.CellSize
	for y := range board.Height {
		for x := range board.Width {
			cell := board.At(sim.Point{X: x, Y: y})
			// One pixel short of the cell size leaves the grid lines showing.
			c.Rect(x*cs, y*cs, cs-1, cs-1, colorFor(pal, cell))
		}
	}
	c.Text(4, cfg.Height*cs+4+textAscent, statusLine(v), colText)

	if v.Standings {
		for i, id := range v.World.Factions() {
			line := fmt.Sprintf("%-10s %5.1f%%", v.World.StrategyName(id), 100*v.World.Shares()[id])
			c.Text(4, cfg.Height*cs+mapHeaderH+i*hudLineHeight+textAscent, line, colText)
		}
	}
}

// DrawBoard composes the standalone leaderboard window onto c.
func DrawBoard(c *canvas.Canvas, state Msg, pal []color.RGBA) {
	c.Fill(background)
	c.Text(6, 6+textAscent, boardHeader(state), colText)

	stats := append([]FactionStat(nil), state.Factions...)
	sort.SliceStable(stats, func(i, j int) bool { return stats[i].Share > stats[j].Share })
	for i, s := range stats {
		y := 6 + (i+2)*hudLineHeight
		c.Rect(6, y+1, 10, 10, pal[(s.ID-1)%len(pal)])
		mark := " "
		if state.Over && s.ID == state.Winner {
			mark = "*"
		}
		c.Text(20, y+textAscent, fmt.Sprintf("%s %-10s %5.1f%%", mark, s.Name, 100*s.Share), colText)
	}
}

// colText is the HUD text colour.
var colText = color.RGBA{R: 230, G: 230, B: 230, A: 255}

// neutralCell is the colour of unclaimed ground.
var neutralCell = color.RGBA{R: 40, G: 40, B: 48, A: 255}

// shadeSteps is how finely a faction's colour is banded by cell strength, and
// so how many shades of it a recording's palette needs.
const shadeSteps = 24

func statusLine(v MapView) string {
	if !v.Over {
		return fmt.Sprintf("tick %d   running", v.World.TickCount())
	}
	if v.Winner != sim.Neutral {
		return fmt.Sprintf("tick %d   winner: %s", v.World.TickCount(), v.World.StrategyName(v.Winner))
	}
	return fmt.Sprintf("tick %d   ended", v.World.TickCount())
}

func boardHeader(state Msg) string {
	if state.Over {
		return fmt.Sprintf("tick %d   ended", state.Tick)
	}
	return fmt.Sprintf("tick %d   running", state.Tick)
}

// MapSize returns the war map's pixel size for the given configuration.
func MapSize(cfg Config, standings bool) (w, h int) {
	h = cfg.Height*cfg.CellSize + mapHeaderH
	if standings {
		h += len(cfg.Strategies) * hudLineHeight
	}
	return cfg.Width * cfg.CellSize, h
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
