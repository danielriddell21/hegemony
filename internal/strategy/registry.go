package strategy

import (
	"strings"

	"github.com/danielriddell21/hegemony/internal/sim"
)

type entry struct {
	name string
	make func() sim.Strategy
}

var registry = []entry{
	{name: "Random", make: Random},
	{name: "Greedy", make: Greedy},
	{name: "Blob", make: Blob},
	{name: "Frontier", make: Frontier},
	{name: "Influence", make: Influence},
	{name: "Bulwark", make: Bulwark},
	{name: "Voronoi", make: Voronoi},
	{name: "Headhunter", make: Headhunter},
	{name: "Lookahead", make: Lookahead},
}

func Names() []string {
	out := make([]string, len(registry))
	for i, e := range registry {
		out[i] = e.name
	}
	return out
}

func New(name string) (sim.Strategy, bool) {
	for _, e := range registry {
		if strings.EqualFold(e.name, name) {
			return e.make(), true
		}
	}
	return nil, false
}
