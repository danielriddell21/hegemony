default:
    @just --list

[group('build')]
build:
    go build ./...

[group('test')]
test:
    go test ./...

[group('dev')]
lint:
    golangci-lint run

[group('dev')]
fmt:
    golangci-lint fmt

[group('dev')]
tidy:
    go mod tidy

[group('dev')]
ci: lint test build

[group('build')]
build-gui:
    go build -tags ebiten -o bin/hegemony ./cmd/hegemony

[group('run')]
gui *ARGS:
    go run -tags ebiten ./cmd/hegemony {{ARGS}}

[group('run')]
run *ARGS:
    go run ./cmd/hegemony headless {{ARGS}}

# regenerate the documentation media
[group('dev')]
demos:
    # Rendered headlessly through the software canvas: no window, no display,
    # no ebiten build tag. A .mp4 path records video instead of a GIF.
    mkdir -p docs/demos
    go run ./cmd/hegemony run --record docs/demos/match.gif --record-frames 200 --seed 3
    go run ./cmd/hegemony run --record docs/demos/duel.gif --record-frames 160 --seed 7 --strategies Blitzkrieg,Turtle,Bulwark
