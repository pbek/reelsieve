package rssfilter

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestIMDBRating(t *testing.T) {
	rating, ok := IMDBRating(`<br />IMDB Rating: 5.6/10<br />Genre: Thriller`)
	if !ok {
		t.Fatal("expected rating")
	}
	if rating != 5.6 {
		t.Fatalf("rating = %v, want 5.6", rating)
	}
}

func TestFilterItems(t *testing.T) {
	items := []Item{
		{Title: "A Movie (2026) [1080p] [WEBRip]", Description: "IMDB Rating: 5.0/10", GUID: "1"},
		{Title: "A Movie (2026) [1080p] [BluRay]", Description: "IMDB Rating: 8.0/10", GUID: "2"},
		{Title: "Bad Movie (2026) [1080p]", Description: "IMDB Rating: 4.9/10", GUID: "3"},
		{Title: "No Rating (2026) [1080p]", Description: "Genre: Drama", GUID: "4"},
		{Title: "Great Movie (2026) [1080p]", Description: "IMDB Rating: 7.1/10", GUID: "5"},
	}

	filtered := FilterItems(items, 5)
	if len(filtered) != 2 {
		t.Fatalf("len(filtered) = %d, want 2", len(filtered))
	}
	if filtered[0].GUID != "1" {
		t.Fatalf("first GUID = %q, want 1", filtered[0].GUID)
	}
	if filtered[1].GUID != "5" {
		t.Fatalf("second GUID = %q, want 5", filtered[1].GUID)
	}
	if !strings.Contains(
		filtered[0].Description,
		`https://www.imdb.com/find/?q=A+Movie+%282026%29&s=tt`,
	) {
		t.Fatalf("description = %q, want IMDb search link", filtered[0].Description)
	}
	if strings.Contains(
		filtered[0].Description,
		`<strong style="color: red; font-weight: bold;">`,
	) {
		t.Fatalf("description = %q, want rating below 7 not highlighted", filtered[0].Description)
	}
	if !strings.Contains(
		filtered[1].Description,
		`IMDB Rating: <strong style="color: red; font-weight: bold;">7.1/10</strong>`,
	) {
		t.Fatalf("description = %q, want highlighted IMDb rating", filtered[1].Description)
	}
}

func TestHighlightIMDBRating(t *testing.T) {
	description := HighlightIMDBRating("IMDB Rating: 7.0/10<br />Genre: Thriller", 7, 7)
	want := `IMDB Rating: <strong style="color: red; font-weight: bold;">7.0/10</strong><br />Genre: Thriller`
	if description != want {
		t.Fatalf("description = %q, want %q", description, want)
	}

	description = HighlightIMDBRating("IMDB Rating: 6.9/10", 6.9, 7)
	if description != "IMDB Rating: 6.9/10" {
		t.Fatalf("description = %q, want unchanged rating below 7", description)
	}

	description = HighlightIMDBRating("IMDB Rating: 6.5/10", 6.5, 6.5)
	want = `IMDB Rating: <strong style="color: red; font-weight: bold;">6.5/10</strong>`
	if description != want {
		t.Fatalf("description = %q, want configurable highlight threshold", description)
	}
}

func TestFilterItemsSkipsStoredItems(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t, 500)
	defer func() {
		if err := store.Close(); err != nil {
			t.Fatalf("Close: %v", err)
		}
	}()

	if err := store.RecordFetched(ctx, []Item{{Title: "Old Movie (2026) [1080p]", GUID: "old"}}); err != nil {
		t.Fatalf("RecordFetched: %v", err)
	}

	service := &Service{minRating: 5, highlightRating: 7, store: store}
	filtered, err := service.filterItems(ctx, []Item{
		{Title: "Old Movie (2026) [BluRay]", Description: "IMDB Rating: 7.0/10", GUID: "1"},
		{Title: "New Movie (2026) [1080p]", Description: "IMDB Rating: 7.1/10", GUID: "2"},
		{Title: "New Movie (2026) [BluRay]", Description: "IMDB Rating: 8.1/10", GUID: "3"},
	})
	if err != nil {
		t.Fatalf("filterItems: %v", err)
	}
	if len(filtered) != 1 {
		t.Fatalf("len(filtered) = %d, want 1", len(filtered))
	}
	if filtered[0].GUID != "2" {
		t.Fatalf("GUID = %q, want 2", filtered[0].GUID)
	}
}

func TestItemStorePrunesToLimit(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t, 2)
	defer func() {
		if err := store.Close(); err != nil {
			t.Fatalf("Close: %v", err)
		}
	}()

	for _, item := range []Item{
		{Title: "One", GUID: "1"},
		{Title: "Two", GUID: "2"},
		{Title: "Three", GUID: "3"},
	} {
		if err := store.RecordFetched(ctx, []Item{item}); err != nil {
			t.Fatalf("RecordFetched: %v", err)
		}
	}

	var count int
	if err := store.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM fetched_items`).Scan(&count); err != nil {
		t.Fatalf("count fetched items: %v", err)
	}
	if count != 2 {
		t.Fatalf("count = %d, want 2", count)
	}
}

func TestNormalizedName(t *testing.T) {
	name := NormalizedName(Item{Title: "Empire of Lies (2026) [1080p] [WEBRip] [x265]"})
	if name != "empire of lies (2026)" {
		t.Fatalf("name = %q, want empire of lies (2026)", name)
	}
}

func TestAddIMDBSearchLink(t *testing.T) {
	item := AddIMDBSearchLink(Item{
		Title:       "Empire of Lies (2026) [1080p] [WEBRip] [x265]",
		Description: `IMDB Rating: <strong style="color: red; font-weight: bold;">7.1/10</strong>`,
	})
	want := `IMDB Rating: <strong style="color: red; font-weight: bold;">7.1/10</strong><br />IMDb: <a href="https://www.imdb.com/find/?q=Empire+of+Lies+%282026%29&s=tt">Search IMDb</a>`
	if item.Description != want {
		t.Fatalf("description = %q, want %q", item.Description, want)
	}

	again := AddIMDBSearchLink(item)
	if again.Description != want {
		t.Fatalf("description = %q, want no duplicate IMDb link", again.Description)
	}
}

func openTestStore(t *testing.T, limit int) *ItemStore {
	t.Helper()
	store, err := OpenItemStore(filepath.Join(t.TempDir(), "items.sqlite3"), limit)
	if err != nil {
		t.Fatalf("OpenItemStore: %v", err)
	}
	return store
}
