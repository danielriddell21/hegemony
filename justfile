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
