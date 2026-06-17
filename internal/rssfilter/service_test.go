package rssfilter

import "testing"

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
}

func TestNormalizedName(t *testing.T) {
	name := NormalizedName(Item{Title: "Empire of Lies (2026) [1080p] [WEBRip] [x265]"})
	if name != "empire of lies (2026)" {
		t.Fatalf("name = %q, want empire of lies (2026)", name)
	}
}
