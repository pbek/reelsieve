# ReelSieve

ReelSieve is a small RSS filtering service for movie feeds. It reads the upstream feed configured by `SOURCE_URL`, keeps items with an IMDB rating of at least `5`, removes duplicate movie names, and serves the result as RSS.

Version: `0.1`

## Run

```sh
docker build --build-arg VERSION=0.1 -t reelsieve:0.1 .
docker run --rm -p 8080:8080 -e SOURCE_URL=https://example.com/feed.xml reelsieve:0.1
```

Open `http://localhost:8080/rss`.

## Configuration

| Variable          | Default  | Description                   |
| ----------------- | -------- | ----------------------------- |
| `LISTEN_ADDR`     | `:8080`  | HTTP listen address           |
| `SOURCE_URL`      | required | Upstream RSS URL              |
| `MIN_RATING`      | `5`      | Minimum IMDB rating           |
| `CACHE_TTL`       | `10m`    | In-memory feed cache duration |
| `REQUEST_TIMEOUT` | `10s`    | Upstream request timeout      |

## Endpoints

| Path       | Description         |
| ---------- | ------------------- |
| `/rss`     | Filtered RSS feed   |
| `/healthz` | Health check        |
| `/version` | Application version |

## Nix

Build the app:

```sh
nix build .#reelsieve
```

Build the container image tarball:

```sh
nix build .#container
```

The flake intentionally exposes build packages only. The local development environment is defined in `devenv.nix` for `devenv shell` / direnv usage and is not exported as a flake dev shell.

## Development

```sh
devenv shell
go test ./...
SOURCE_URL=https://example.com/feed.xml go run ./cmd/reelsieve
```
