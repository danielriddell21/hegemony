package evolve_test

import (
	"testing"

	"github.com/danielriddell21/hegemony/internal/evolve"
	"github.com/danielriddell21/hegemony/internal/sim"
	"github.com/danielriddell21/hegemony/internal/strategy"
)

func smallConfig() evolve.Config {
	return evolve.Config{
		Width: 12, Height: 12, Seeds: 4, Ticks: 150, Threshold: 0,
		Params:      sim.DefaultParams(),
		Opponents:   []sim.Strategy{strategy.Greedy(), strategy.Blob()},
		Population:  12,
		Generations: 8,
		Seed:        1,
	}
}

func TestRunNeverRegresses(t *testing.T) {
	res := evolve.Run(smallConfig())
	if res.Fitness < res.Baseline {
		t.Errorf("evolution regressed: fitness %.4f < baseline %.4f", res.Fitness, res.Baseline)
	}
	if res.Fitness < 0 || res.Fitness > 1 {
		t.Errorf("fitness %.4f out of [0,1]", res.Fitness)
	}
}

func TestRunReproducible(t *testing.T) {
	a := evolve.Run(smallConfig())
	b := evolve.Run(smallConfig())
	if a != b {
		t.Errorf("evolution not reproducible:\n%+v\n%+v", a, b)
	}
}
