package tournament

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/danielriddell21/hegemony/internal/sim"
	"github.com/danielriddell21/hegemony/internal/strategy"
)

type Config struct {
	Width        int
	Height       int
	Seeds        int
	BaseSeed     uint64
	MaxTicks     int
	WinThreshold float64
	Params       sim.Params
	Strategies   []string
}

type Standing struct {
	Name         string
	Matches      int
	Wins         int
	TerritorySum float64
}

func (s Standing) WinRate() float64 {
	if s.Matches == 0 {
		return 0
	}
	return float64(s.Wins) / float64(s.Matches)
}

func (s Standing) MeanTerritory() float64 {
	if s.Matches == 0 {
		return 0
	}
	return s.TerritorySum / float64(s.Matches)
}

func Run(cfg Config) ([]Standing, error) {
	if len(cfg.Strategies) < 2 {
		return nil, fmt.Errorf("tournament needs at least 2 strategies, got %d", len(cfg.Strategies))
	}
	if cfg.Seeds < 1 {
		return nil, fmt.Errorf("tournament needs at least 1 seed, got %d", cfg.Seeds)
	}

	entrants, err := resolve(cfg.Strategies)
	if err != nil {
		return nil, err
	}
	spawns := sim.SpreadSpawns(cfg.Width, cfg.Height, len(entrants))

	standings := make([]Standing, len(entrants))
	for i, s := range entrants {
		standings[i] = Standing{Name: s.Name()}
	}

	total := float64(cfg.Width * cfg.Height)
	for s := range cfg.Seeds {
		res := sim.RunMatch(matchConfig(cfg, entrants, spawns, cfg.BaseSeed+uint64(s)))
		for i := range standings {
			id := sim.FactionID(i + 1)
			standings[i].Matches++
			standings[i].TerritorySum += float64(res.Territory[id]) / total
			if res.Winner == id {
				standings[i].Wins++
			}
		}
	}

	slices.SortFunc(standings, func(a, b Standing) int {
		if c := cmp.Compare(b.WinRate(), a.WinRate()); c != 0 {
			return c
		}
		if c := cmp.Compare(b.MeanTerritory(), a.MeanTerritory()); c != 0 {
			return c
		}
		return cmp.Compare(a.Name, b.Name)
	})
	return standings, nil
}

func resolve(names []string) ([]sim.Strategy, error) {
	out := make([]sim.Strategy, 0, len(names))
	for _, name := range names {
		s, ok := strategy.New(name)
		if !ok {
			return nil, fmt.Errorf("unknown strategy %q", name)
		}
		out = append(out, s)
	}
	return out, nil
}

func matchConfig(cfg Config, entrants []sim.Strategy, spawns []sim.Point, seed uint64) sim.MatchConfig {
	list := make([]sim.Entrant, len(entrants))
	for i, s := range entrants {
		list[i] = sim.Entrant{Strategy: s, Spawn: spawns[i]}
	}
	return sim.MatchConfig{
		Width:        cfg.Width,
		Height:       cfg.Height,
		Seed:         seed,
		MaxTicks:     cfg.MaxTicks,
		WinThreshold: cfg.WinThreshold,
		Params:       cfg.Params,
		Entrants:     list,
	}
}
