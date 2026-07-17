package gui

// Msg is the line-delimited JSON envelope exchanged between the leader and its
// child windows. Only the war-map leader produces "state" messages; the
// leaderboard consumes them. "quit" asks a window to close.
type Msg struct {
	Type     string        `json:"t"`
	Tick     int           `json:"tick,omitempty"`
	Over     bool          `json:"over,omitempty"`
	Winner   int           `json:"winner,omitempty"`
	Factions []FactionStat `json:"f,omitempty"`
}

// FactionStat is one row of the shared leaderboard: a faction's id (which also
// selects its colour), strategy name, and current territory share.
type FactionStat struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Share float64 `json:"share"`
}

// Link is a window's end of the coordination channel: state it receives on In,
// state it publishes on Out. A nil Link means a standalone, single window.
type Link struct {
	In  <-chan Msg
	Out chan<- Msg
}
