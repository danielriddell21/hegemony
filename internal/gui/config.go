package gui

import (
	"github.com/danielriddell21/crucible/record"

	"github.com/danielriddell21/hegemony/internal/sim"
)

type Role string

const (
	RoleMap   Role = "map"
	RoleBoard Role = "board"
)

type Config struct {
	Width          int
	Height         int
	Seed           uint64
	MaxTicks       int
	WinThreshold   float64
	Params         sim.Params
	Strategies     []string
	CellSize       int
	TicksPerSecond int

	Role        Role
	OffsetIndex int
	Link        *Link

	// Rec holds the shared --record flags; when its path is set the war map
	// runs standalone, captures frames to a GIF and exits.
	Rec record.Options
}
