import ".shared/common.just"

version := `cat internal/version/VERSION`
image := "reelsieve:" + version

default:
    just --list

test:
    go test ./...

build:
    go build -trimpath -ldflags="-s -w" -o ./bin/reelsieve ./cmd/reelsieve

run:
    go run ./cmd/reelsieve

docker-build:
    docker build -t {{ image }} .

docker-run: docker-build
    docker run --rm -p 8080:8080 {{ image }}

nix-build:
    nix build .#reelsieve

nix-container:
    nix build .#container
