import ".shared/common.just"

version := "0.1"
image := "reelsieve:" + version

default:
    just --list

test:
    go test ./...

build:
    go build -trimpath -ldflags="-s -w -X main.version={{ version }}" -o ./bin/reelsieve ./cmd/reelsieve

run:
    go run -ldflags="-X main.version={{ version }}" ./cmd/reelsieve

docker-build:
    docker build --build-arg VERSION={{ version }} -t {{ image }} .

docker-run: docker-build
    docker run --rm -p 8080:8080 {{ image }}

nix-build:
    nix build .#reelsieve

nix-container:
    nix build .#container
