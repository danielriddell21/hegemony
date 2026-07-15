package gui

import "github.com/danielriddell21/hegemony/internal/sim"

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
}
