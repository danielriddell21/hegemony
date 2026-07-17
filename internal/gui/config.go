package gui

import "github.com/danielriddell21/hegemony/internal/sim"

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
}
