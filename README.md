# hegemony

A territory-war simulation where competing algorithms fight to control a shared
grid. Several factions start from spawn cells and expand, attack, and defend to
claim the map. Each faction is driven by a pluggable **strategy** — the whole
point is to pit strategies against each other and see which one controls the
most territory.

## Install

```sh
go install github.com/danielriddell21/hegemony/cmd/hegemony@latest
```

Or build from a checkout with [`just`](https://just.systems):

```sh
just build          # headless binary
just build-gui      # with the Ebiten visualizer (-tags ebiten)
```

## Usage

### Tournament (headless)

Run a free-for-all across the strategy roster over many seeds and print a
reproducible win-rate / mean-territory table:

```sh
hegemony headless --seeds 50 --width 24 --height 24
```

```
Strategy      Matches   Wins   WinRate  MeanTerritory
Greedy             50     50    100.0%          60.9%
Blob               50      0      0.0%          32.7%
Random             50      0      0.0%           0.8%
```

Pick specific entrants with `--strategies Greedy,Blob`. Every run is seeded, so
the same flags always produce the same table.

### Watch a match (GUI)

Build with `-tags ebiten` to watch a single match — the grid is colored by
owner, with a HUD showing the tick and each faction's territory share:

```sh
just gui run --width 32 --height 32 --strategies Greedy,Blob,Random
```

Without the `ebiten` build tag the `run` command reports that the GUI is not
compiled in; the headless simulation needs no GUI.

## How it works

- The map is a grid of cells; each cell is neutral or owned by one faction and
  carries a **strength** value.
- Every tick each owned cell gains income as **strength** (`--income` per cell,
  capped at `--max-strength`), so a cell's garrison is both its defence and its
  attacking reserve.
- An **action** is a troop move: send some strength from an owned cell into an
  adjacent cell. Into your own cell it reinforces; into a neutral or enemy cell
  it attacks. The simulation validates and applies actions — strategies never
  mutate the world directly.
- Contests (moved strength vs. the defender's, with seeded jitter) are resolved
  deterministically, so the same seed always replays the same match.
- A match ends when a faction controls a threshold share of the map, when only
  one faction remains, or when the tick limit is reached (most territory wins).

Because strength is the currency for both offence and defence, the rules are
sharply tunable: `--max-strength`, `--income`, `--neutral-defense`,
`--start-strength`, and `--jitter` swing the meta between wall-favouring and
spear-favouring — walls (`Turtle`) beat spears at high `--max-strength`, spears
(`Blitzkrieg`) break walls when it is low.

## Strategies

A strategy implements a single interface:

```go
type Strategy interface {
	Name() string
	Move(v View, rng *rand.Rand) []Action
}
```

`View` is a read-only snapshot of the board — ownership, cell strength, and the
acting faction's owned cells and frontier. Each strategy decides how much
strength to move from which owned cells and where. The roster:

| Strategy | Idea |
| --- | --- |
| `Random` | Push random amounts from random border cells onto random frontier cells. |
| `Greedy` | Grab the weakest adjacent cells first, capturing as many as strength allows. |
| `Blob` | Flood outward, pushing on the whole frontier at once. |
| `Frontier` | Expand toward open ground — weight each capture by open space unlocked ÷ cost. |
| `Influence` | Capture where your local presence dominates the enemy's, avoiding overextension. |
| `Bulwark` | Expand, then fortify only the cells that touch an enemy, denying cheap counter-captures. |
| `Voronoi` | Claim the hinterland far from any enemy first, locking in the larger region. |
| `Headhunter` | Drive toward the smallest surviving faction and eat it. |
| `Turtle` | Take only the cheapest ground and pour everything else into an impregnable perimeter. |
| `Blitzkrieg` | Concentrate a whole spearhead cell into one overwhelming strike and roll it forward. |
| `Lookahead` | Meta-strategy: simulate several strategies' moves one step in a sandbox and play the best-scoring one. |

The strategies are strongly non-transitive — which one controls the most map
depends on the field, the board, and the rules. Under the default (defensive)
economy, fortifying strategies like `Bulwark` and `Turtle` rise while thin
expanders get rolled back; dial `--max-strength` down and `Blitzkrieg`'s spears
start cracking the walls. There is no single best strategy, which is the whole
point — the flags let you find the meta you want.

New strategies register in `internal/strategy` and are picked up by both the
tournament and the GUI automatically.

## Development

```sh
just ci             # lint + test + build
just test           # go test ./...
just lint           # golangci-lint run
```

See [CONVENTIONS.md](CONVENTIONS.md) for the CLI and GUI structure shared across
the tool family, and [CLAUDE.md](CLAUDE.md) for contributor guidance.

## License

[MIT](LICENSE) © Dan Riddell
