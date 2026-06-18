# Changelog

## 0.2 - 2026-06-18

- Add IMDb search links to filtered RSS item descriptions.
- Use `internal/version/VERSION` as the single release version source for the app and build tooling.
- Use GoReleaser to publish GitHub Releases with changelog release notes from pushes to the `release` branch.
- Keep duplicate prevention and rating filtering behavior unchanged.

## 0.1 - 2026-06-17

- Initial RSS filtering service.
- Filter movie feed items by minimum IMDb rating.
- Store fetched item keys in SQLite to avoid resurfacing previously seen movies.
- Remove duplicate movie names within each fetched feed.
