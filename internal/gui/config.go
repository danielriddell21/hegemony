package gui

import "github.com/danielriddell21/hegemony/internal/sim"

// Role selects which window a process renders.
type Role string

const (
	// RoleMap is the leader window: it owns the simulation, draws the war map,
	// and publishes the shared leaderboard state.
	RoleMap Role = "map"
	// RoleBoard is the child window: it draws the leaderboard from received
	// state and runs no simulation.
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
