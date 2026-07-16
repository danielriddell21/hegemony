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
Search             50     34     68.0%          52.1%
Bulwark            50     16     32.0%          42.8%
Blitzkrieg         50      0      0.0%           2.9%
Turtle             50      0      0.0%           1.3%
...
```

Pick specific entrants with `--strategies Search,Bulwark`. Every run is seeded,
so the same flags always produce the same table.

### Watch a match (GUI)

Build with `-tags ebiten` to watch a single match — the grid is colored by
owner, with a HUD showing the tick and each faction's territory share:

```sh
just gui run --width 32 --height 32 --strategies Greedy,Blob,Random
```

Without the `ebiten` build tag the `run` command reports that the GUI is not
compiled in; the headless simulation needs no GUI.

### Evolve a strategy

The `Evolved` strategy scores frontier cells with a weighted blend of the other
strategies' features. `hegemony evolve` hill-climbs those weights offline,
scoring each candidate by the territory it holds in one-on-one duels against a
panel of opponents, and prints the best vector (paste it into
`internal/strategy/evolved.go`):

```sh
hegemony evolve --width 20 --height 20 --seeds 16 --generations 80 --seed 1
```

Everything is seeded, so the search is reproducible. The baked-in weights make
`Evolved` competitive with the hand-written expanders one-on-one (46–50% of the
map against each) — it learned a counter-intuitive policy, pushing *toward* the
enemy into open ground rather than copying any single hand-coded heuristic.

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
  deterministically, so the same seed always replays the same match. Faction
  move order is reshuffled every tick so no one has a fixed first-mover edge.
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
| `Evolved` | A linear blend of the features above whose weights were tuned offline by `hegemony evolve`. |
| `Lookahead` | Meta-strategy: simulate several strategies' moves one step in a sandbox and play the best-scoring one. |
| `Search` | Forks the live position into a forward model, projects each candidate strategy several ticks ahead, and commits to the one that ends up holding the most map. |

Which one controls the most map depends on the field, the board, and the rules.
Under the default (defensive) economy the forward-model `Search` edges out the
fortifying `Bulwark`, while thin expanders get rolled back; dial `--max-strength`
down and `Blitzkrieg`'s spears start cracking the walls. The roster is strongly
non-transitive — no single strategy dominates every setting, which is the whole
point — and the flags let you find the meta you want.

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
