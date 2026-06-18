# Changelog

## 0.2 - 2026-06-18

- Add IMDb search links to filtered RSS item descriptions.
- Keep duplicate prevention and rating filtering behavior unchanged.

## 0.1 - 2026-06-17

- Initial RSS filtering service.
- Filter movie feed items by minimum IMDb rating.
- Store fetched item keys in SQLite to avoid resurfacing previously seen movies.
- Remove duplicate movie names within each fetched feed.
