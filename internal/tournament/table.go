package tournament

import (
	"fmt"
	"strings"
)

func Table(standings []Standing) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%-12s %8s %6s %9s %14s\n", "Strategy", "Matches", "Wins", "WinRate", "MeanTerritory")
	for _, s := range standings {
		fmt.Fprintf(&b, "%-12s %8d %6d %8.1f%% %13.1f%%\n",
			s.Name, s.Matches, s.Wins, 100*s.WinRate(), 100*s.MeanTerritory())
	}
	return b.String()
}
