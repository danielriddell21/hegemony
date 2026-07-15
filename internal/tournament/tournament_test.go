package tournament_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/danielriddell21/hegemony/internal/sim"
	"github.com/danielriddell21/hegemony/internal/tournament"
)

func baseConfig() tournament.Config {
	return tournament.Config{
		Width: 16, Height: 16, Seeds: 12, BaseSeed: 1,
		MaxTicks: 300, WinThreshold: 0.6,
		Params:     sim.DefaultParams(),
		Strategies: []string{"Greedy", "Blob", "Random"},
	}
}

func TestRunTalliesMatches(t *testing.T) {
	standings, err := tournament.Run(baseConfig())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(standings) != 3 {
		t.Fatalf("got %d standings, want 3", len(standings))
	}
	totalWins := 0
	for _, s := range standings {
		if s.Matches != 12 {
			t.Errorf("%s matches = %d, want 12", s.Name, s.Matches)
		}
		totalWins += s.Wins
	}
	if totalWins > 12 {
		t.Errorf("total wins %d exceeds match count 12", totalWins)
	}
	// Sorted by win rate descending.
	if standings[0].WinRate() < standings[len(standings)-1].WinRate() {
		t.Error("standings not sorted by win rate descending")
	}
}

func TestRunReproducible(t *testing.T) {
	a, err := tournament.Run(baseConfig())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	b, err := tournament.Run(baseConfig())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Fatal("tournament not reproducible across runs")
	}
}

func TestRunValidatesInput(t *testing.T) {
	cfg := baseConfig()
	cfg.Strategies = []string{"Greedy"}
	if _, err := tournament.Run(cfg); err == nil {
		t.Error("expected error for fewer than 2 strategies")
	}

	cfg = baseConfig()
	cfg.Strategies = []string{"Greedy", "Ghost"}
	if _, err := tournament.Run(cfg); err == nil {
		t.Error("expected error for unknown strategy")
	}
}

func TestTableRendersRows(t *testing.T) {
	standings, err := tournament.Run(baseConfig())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	table := tournament.Table(standings)
	if !strings.Contains(table, "Strategy") || !strings.Contains(table, "WinRate") {
		t.Error("table missing header")
	}
	for _, s := range standings {
		if !strings.Contains(table, s.Name) {
			t.Errorf("table missing row for %s", s.Name)
		}
	}
}
