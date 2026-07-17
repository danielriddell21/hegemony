package gui

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

type Link struct {
	In  <-chan Msg
	Out chan<- Msg
}
