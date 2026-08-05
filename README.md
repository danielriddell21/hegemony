# hegemony

[![CI](https://github.com/danielriddell21/hegemony/actions/workflows/ci.yaml/badge.svg)](https://github.com/danielriddell21/hegemony/actions/workflows/ci.yaml)
[![codecov](https://codecov.io/gh/danielriddell21/hegemony/branch/trunk/graph/badge.svg)](https://codecov.io/gh/danielriddell21/hegemony)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=danielriddell21_hegemony&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=danielriddell21_hegemony)
[![Go 1.26](https://img.shields.io/badge/go-1.26-blue)](https://go.dev)
[![MIT License](https://img.shields.io/badge/licence-MIT-green)](LICENSE)

A territory-war simulation where competing algorithms fight to control a shared grid.

## Commands

| Command | Description |
|---|---|
| `hegemony headless` | Run a batch tournament across strategies and print a result table |
| `hegemony run` | Watch a match in the GUI — war map and leaderboard in separate windows |
| `hegemony evolve` | Search weights for the Evolved strategy and print the best vector |

## Install

### Homebrew
```bash
# headless CLI
brew install danielriddell21/tap/hegemony

# GUI
brew install --cask danielriddell21/tap/hegemony
```

### Go install
```bash
go install github.com/danielriddell21/hegemony/cmd/hegemony@latest
```

### From source
```bash
git clone https://github.com/danielriddell21/hegemony
cd hegemony
just gui run
```

A `go install` build is headless only — the GUI sits behind the `ebiten` build tag.

## Documentation

Full documentation lives in the [hegemony wiki](https://github.com/danielriddell21/hegemony/wiki).
