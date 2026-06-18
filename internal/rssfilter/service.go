package rssfilter

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pbek/reelsieve/internal/version"
)

var (
	ratingPattern = regexp.MustCompile(`(?i)IMDB\s+Rating:\s*([0-9]+(?:\.[0-9]+)?)/10`)
	bracketGroup  = regexp.MustCompile(`\s*\[[^\]]+\]`)
	spacePattern  = regexp.MustCompile(`\s+`)
)

type Service struct {
	client    *http.Client
	sourceURL string
	minRating float64
	cacheTTL  time.Duration
	store     *ItemStore

	mu        sync.RWMutex
	cached    RSS
	expiresAt time.Time
}

func NewService(
	client *http.Client,
	sourceURL string,
	minRating float64,
	cacheTTL time.Duration,
	store *ItemStore,
) *Service {
	return &Service{
		client:    client,
		sourceURL: sourceURL,
		minRating: minRating,
		cacheTTL:  cacheTTL,
		store:     store,
	}
}

func (s *Service) Feed(ctx context.Context) (RSS, error) {
	now := time.Now()
	s.mu.RLock()
	if !s.expiresAt.IsZero() && now.Before(s.expiresAt) {
		feed := s.cached
		s.mu.RUnlock()
		return feed, nil
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.expiresAt.IsZero() && time.Now().Before(s.expiresAt) {
		return s.cached, nil
	}

	feed, err := s.fetch(ctx)
	if err != nil {
		return RSS{}, err
	}
	fetchedItems := feed.Channel.Items
	filteredItems, err := s.filterItems(ctx, fetchedItems)
	if err != nil {
		return RSS{}, err
	}
	feed.Channel.Items = filteredItems
	if s.store != nil {
		if err := s.store.RecordFetched(ctx, fetchedItems); err != nil {
			return RSS{}, err
		}
	}
	if feed.Channel.Title != "" {
		feed.Channel.Title = "ReelSieve - " + feed.Channel.Title
	}
	if feed.Channel.AtomLink != nil {
		feed.Channel.AtomLink = nil
	}

	s.cached = feed
	s.expiresAt = time.Now().Add(s.cacheTTL)
	return feed, nil
}

func (s *Service) fetch(ctx context.Context) (RSS, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.sourceURL, nil)
	if err != nil {
		return RSS{}, err
	}
	req.Header.Set("User-Agent", "reelsieve/"+version.String())
	req.Header.Set("Accept", "application/rss+xml, application/xml;q=0.9, text/xml;q=0.8")

	res, err := s.client.Do(req)
	if err != nil {
		return RSS{}, err
	}
	defer func() {
		_ = res.Body.Close()
	}()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, res.Body)
		return RSS{}, fmt.Errorf("upstream returned %s", res.Status)
	}

	var feed RSS
	decoder := xml.NewDecoder(io.LimitReader(res.Body, 4<<20))
	if err := decoder.Decode(&feed); err != nil {
		return RSS{}, err
	}
	if feed.Version == "" {
		feed.Version = "2.0"
	}
	return feed, nil
}

func (s *Service) filterItems(ctx context.Context, items []Item) ([]Item, error) {
	filtered := make([]Item, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		rating, ok := IMDBRating(item.Description)
		if !ok || rating < s.minRating {
			continue
		}

		key := ItemKey(item)
		if _, exists := seen[key]; exists {
			continue
		}
		if s.store != nil {
			exists, err := s.store.Seen(ctx, key)
			if err != nil {
				return nil, err
			}
			if exists {
				continue
			}
		}
		seen[key] = struct{}{}
		item = AddIMDBSearchLink(item)
		filtered = append(filtered, item)
	}
	return filtered, nil
}

func FilterItems(items []Item, minRating float64) []Item {
	filtered := make([]Item, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		rating, ok := IMDBRating(item.Description)
		if !ok || rating < minRating {
			continue
		}

		key := ItemKey(item)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		item = AddIMDBSearchLink(item)
		filtered = append(filtered, item)
	}
	return filtered
}

func AddIMDBSearchLink(item Item) Item {
	query := SearchTitle(item)
	if query == "" || strings.Contains(item.Description, "imdb.com/find/") {
		return item
	}

	link := "https://www.imdb.com/find/?q=" + url.QueryEscape(query) + "&s=tt"
	separator := ""
	if strings.TrimSpace(item.Description) != "" {
		separator = "<br />"
	}
	item.Description += separator + `IMDb: <a href="` + link + `">Search IMDb</a>`
	return item
}

func SearchTitle(item Item) string {
	title := strings.TrimSpace(item.Title)
	if title == "" {
		title = strings.TrimSpace(item.Link)
	}
	title = bracketGroup.ReplaceAllString(title, "")
	title = spacePattern.ReplaceAllString(title, " ")
	return strings.TrimSpace(title)
}

func ItemKey(item Item) string {
	name := NormalizedName(item)
	if name == "" {
		name = strings.ToLower(strings.TrimSpace(item.GUID))
	}
	return name
}

func IMDBRating(description string) (float64, bool) {
	matches := ratingPattern.FindStringSubmatch(description)
	if len(matches) != 2 {
		return 0, false
	}
	rating, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, false
	}
	return rating, true
}

func NormalizedName(item Item) string {
	title := strings.TrimSpace(item.Title)
	if title == "" {
		title = strings.TrimSpace(item.Link)
	}
	name := bracketGroup.ReplaceAllString(title, "")
	name = strings.ToLower(spacePattern.ReplaceAllString(name, " "))
	return strings.TrimSpace(name)
}
