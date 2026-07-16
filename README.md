# hegemony

A territory-war simulation where competing algorithms fight to control a shared grid.

[![CI](https://github.com/danielriddell21/hegemony/actions/workflows/ci.yaml/badge.svg)](https://github.com/danielriddell21/hegemony/actions/workflows/ci.yaml)
[![codecov](https://codecov.io/gh/danielriddell21/hegemony/branch/trunk/graph/badge.svg)](https://codecov.io/gh/danielriddell21/hegemony)
[![Quality Gate](https://sonarcloud.io/api/project_badges/measure?project=danielriddell21_hegemony&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=danielriddell21_hegemony)
[![Go](https://img.shields.io/github/go-mod/go-version/danielriddell21/hegemony)](go.mod)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## Install

```sh
go install github.com/danielriddell21/hegemony/cmd/hegemony@latest
```

## Usage

Run a headless tournament across the strategy roster:

```sh
hegemony headless --seeds 50 --width 24 --height 24
```

Watch a single match in the GUI (built with `-tags ebiten`):

```sh
just gui run --width 32 --height 32
```
