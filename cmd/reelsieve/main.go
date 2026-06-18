package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/pbek/reelsieve/internal/rssfilter"
	appversion "github.com/pbek/reelsieve/internal/version"
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	client := &http.Client{Timeout: cfg.requestTimeout}
	store, err := rssfilter.OpenItemStore(cfg.fetchedItemsDBPath, cfg.fetchedItemsLimit)
	if err != nil {
		slog.Error("failed to open fetched item database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := store.Close(); err != nil {
			slog.Error("failed to close fetched item database", "error", err)
		}
	}()
	service := rssfilter.NewService(client, cfg.sourceURL, cfg.minRating, cfg.cacheTTL, store)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /rss", func(w http.ResponseWriter, r *http.Request) {
		feed, err := service.Feed(r.Context())
		if err != nil {
			slog.Error("failed to render feed", "error", err)
			http.Error(w, "failed to render feed", http.StatusBadGateway)
			return
		}

		w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(xml.Header))
		if err := xml.NewEncoder(w).Encode(feed); err != nil {
			slog.Error("failed to encode feed", "error", err)
		}
	})
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /version", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = fmt.Fprintln(w, appversion.String())
	})
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/rss", http.StatusFound)
	})

	server := &http.Server{
		Addr:              cfg.listenAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	slog.Info(
		"starting reelsieve",
		"version",
		appversion.String(),
		"addr",
		cfg.listenAddr,
		"source",
		cfg.sourceURL,
		"min_rating",
		cfg.minRating,
		"fetched_items_db",
		cfg.fetchedItemsDBPath,
		"fetched_items_limit",
		cfg.fetchedItemsLimit,
	)
	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

type config struct {
	listenAddr         string
	sourceURL          string
	minRating          float64
	cacheTTL           time.Duration
	requestTimeout     time.Duration
	fetchedItemsDBPath string
	fetchedItemsLimit  int
}

func loadConfig() (config, error) {
	cfg := config{
		listenAddr:         getEnv("LISTEN_ADDR", ":8080"),
		minRating:          5,
		cacheTTL:           10 * time.Minute,
		requestTimeout:     10 * time.Second,
		fetchedItemsDBPath: getEnv("FETCHED_ITEMS_DB_PATH", "reelsieve.sqlite3"),
		fetchedItemsLimit:  500,
	}

	cfg.sourceURL = os.Getenv("SOURCE_URL")
	if cfg.sourceURL == "" {
		return cfg, errors.New("SOURCE_URL is required")
	}

	if value := os.Getenv("MIN_RATING"); value != "" {
		parsed, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return cfg, fmt.Errorf("MIN_RATING: %w", err)
		}
		cfg.minRating = parsed
	}
	if value := os.Getenv("CACHE_TTL"); value != "" {
		parsed, err := time.ParseDuration(value)
		if err != nil {
			return cfg, fmt.Errorf("CACHE_TTL: %w", err)
		}
		cfg.cacheTTL = parsed
	}
	if value := os.Getenv("REQUEST_TIMEOUT"); value != "" {
		parsed, err := time.ParseDuration(value)
		if err != nil {
			return cfg, fmt.Errorf("REQUEST_TIMEOUT: %w", err)
		}
		cfg.requestTimeout = parsed
	}
	if value := os.Getenv("FETCHED_ITEMS_LIMIT"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return cfg, fmt.Errorf("FETCHED_ITEMS_LIMIT: %w", err)
		}
		cfg.fetchedItemsLimit = parsed
	}
	if cfg.fetchedItemsLimit <= 0 {
		return cfg, errors.New("FETCHED_ITEMS_LIMIT must be greater than zero")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
