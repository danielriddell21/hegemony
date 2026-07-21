package gui

import "github.com/danielriddell21/crucible/hub"

type Msg struct {
	Type     string        `json:"t"`
	Tick     int           `json:"tick,omitempty"`
	Over     bool          `json:"over,omitempty"`
	Winner   int           `json:"winner,omitempty"`
	Factions []FactionStat `json:"f,omitempty"`
}

type FactionStat struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Share float64 `json:"share"`
}

// Link is the window's channel pair to the multi-window hub, aliasing the
// engine's generic hub link specialised to Msg.
type Link = hub.Link[Msg]
