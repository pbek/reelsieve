package main

import "testing"

func TestLoadConfigHighlightRating(t *testing.T) {
	t.Setenv("SOURCE_URL", "https://example.com/feed.xml")
	t.Setenv("HIGHLIGHT_RATING", "6.5")

	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if cfg.highlightRating != 6.5 {
		t.Fatalf("highlightRating = %v, want 6.5", cfg.highlightRating)
	}
}
