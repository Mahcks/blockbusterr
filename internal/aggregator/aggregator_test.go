package aggregator

import (
	"testing"
	"time"

	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/scoring"
)

func TestShouldUseGlobalLimits(t *testing.T) {
	cfg := &config.Config{}
	cfg.Jobs.GlobalLimitMovies = 10
	cfg.Jobs.GlobalLimitShows = 0

	agg := New(cfg, nil)

	if !agg.ShouldUseGlobalLimits("movie") {
		t.Error("should use global limits for movies")
	}

	if agg.ShouldUseGlobalLimits("show") {
		t.Error("should not use global limits for shows")
	}
}

func TestRankAndLimit(t *testing.T) {
	cfg := &config.Config{}
	cfg.Jobs.GlobalLimitMovies = 3
	cfg.Jobs.GlobalPeriod = "sync"

	agg := New(cfg, nil)

	items := []scoring.ContentScore{
		{Title: "Movie A", Score: 0.95, MediaType: "movie"},
		{Title: "Movie B", Score: 0.85, MediaType: "movie"},
		{Title: "Movie C", Score: 0.75, MediaType: "movie"},
		{Title: "Movie D", Score: 0.65, MediaType: "movie"},
		{Title: "Movie E", Score: 0.55, MediaType: "movie"},
	}

	result := agg.RankAndLimit(items, "movie")

	// Should return only top 3
	if len(result) != 3 {
		t.Errorf("expected 3 items, got %d", len(result))
	}

	// Should be sorted by score
	for i := 0; i < len(result)-1; i++ {
		if result[i].Score < result[i+1].Score {
			t.Errorf("items not sorted by score")
		}
	}

	// Check ranks
	for i, item := range result {
		if item.Rank != i+1 {
			t.Errorf("item %d has rank %d, expected %d", i, item.Rank, i+1)
		}
	}

	// Top item should be highest scored
	if result[0].Title != "Movie A" {
		t.Errorf("top item should be Movie A, got %s", result[0].Title)
	}

	t.Logf("Selected top 3: %s (%.2f), %s (%.2f), %s (%.2f)",
		result[0].Title, result[0].Score,
		result[1].Title, result[1].Score,
		result[2].Title, result[2].Score)
}

func TestRankAndLimitNoLimit(t *testing.T) {
	cfg := &config.Config{}
	cfg.Jobs.GlobalLimitMovies = 0
	cfg.Jobs.GlobalPeriod = "sync"

	agg := New(cfg, nil)

	items := []scoring.ContentScore{
		{Title: "Movie A", Score: 0.95, MediaType: "movie"},
		{Title: "Movie B", Score: 0.85, MediaType: "movie"},
		{Title: "Movie C", Score: 0.75, MediaType: "movie"},
	}

	result := agg.RankAndLimit(items, "movie")

	// Should return all items when no limit
	if len(result) != len(items) {
		t.Errorf("expected %d items, got %d", len(items), len(result))
	}
}

func TestApplyMinimumPicks(t *testing.T) {
	cfg := &config.Config{}
	cfg.Jobs.GlobalLimitMovies = 7
	cfg.Jobs.GlobalPeriod = "sync"

	agg := New(cfg, nil)

	items := []scoring.ContentScore{
		{Title: "Popular A", Score: 0.95, MediaType: "movie", JobName: "popular_movies"},
		{Title: "Popular B", Score: 0.93, MediaType: "movie", JobName: "popular_movies"},
		{Title: "BoxOffice C", Score: 0.91, MediaType: "movie", JobName: "box_office"},
		{Title: "Popular D", Score: 0.89, MediaType: "movie", JobName: "popular_movies"},
		{Title: "Popular E", Score: 0.88, MediaType: "movie", JobName: "popular_movies"},
		{Title: "BoxOffice F", Score: 0.87, MediaType: "movie", JobName: "box_office"},
		{Title: "Popular G", Score: 0.86, MediaType: "movie", JobName: "popular_movies"},
		{Title: "Trending H", Score: 0.75, MediaType: "movie", JobName: "trending_movies"}, // Below top 7
	}

	// Sort items by score (as they would be in real usage)
	ranked := make([]scoring.ContentScore, len(items))
	copy(ranked, items)
	// Already sorted by score in this test data

	// Apply minimum: trending must have at least 1 pick
	jobMinimums := []JobMinimum{
		{JobName: "trending_movies", Minimum: 1},
	}

	// Pass ALL items, not just top 7 - the function will select top 7 with minimums applied
	result := agg.ApplyMinimumPicks(ranked, jobMinimums, 7)

	// Should still have 7 items
	if len(result) != 7 {
		t.Errorf("expected 7 items, got %d", len(result))
	}

	// Count job picks
	trendingCount := 0
	for _, item := range result {
		if item.JobName == "trending_movies" {
			trendingCount++
		}
	}

	// Trending should have at least 1 pick
	if trendingCount < 1 {
		t.Errorf("trending should have at least 1 pick, got %d", trendingCount)
	}

	// Trending H should be in the results
	found := false
	for _, item := range result {
		if item.Title == "Trending H" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Trending H should be in results to meet minimum requirement")
	}

	t.Logf("With minimums - Trending picks: %d/%d", trendingCount, len(result))
}

func TestApplyMinimumPicksNoMinimums(t *testing.T) {
	cfg := &config.Config{}
	cfg.Jobs.GlobalLimitMovies = 5

	agg := New(cfg, nil)

	items := []scoring.ContentScore{
		{Title: "Movie A", Score: 0.95, MediaType: "movie", JobName: "popular_movies"},
		{Title: "Movie B", Score: 0.85, MediaType: "movie", JobName: "popular_movies"},
		{Title: "Movie C", Score: 0.75, MediaType: "movie", JobName: "trending_movies"},
	}

	// No minimums configured
	result := agg.ApplyMinimumPicks(items, []JobMinimum{}, 5)

	// Should return all items (since we have fewer than limit)
	if len(result) != len(items) {
		t.Errorf("expected %d items, got %d", len(items), len(result))
	}
}

func TestCalculateResetTimeSync(t *testing.T) {
	cfg := &config.Config{}
	agg := New(cfg, nil)

	now := time.Now()
	resetTime := agg.CalculateResetTime("sync")

	// Sync should reset immediately (at or after current time)
	if resetTime.Before(now) {
		t.Error("sync reset time should not be in the past")
	}
}

func TestCalculateResetTimeDaily(t *testing.T) {
	cfg := &config.Config{}
	agg := New(cfg, nil)

	now := time.Now()
	resetTime := agg.CalculateResetTime("daily")

	// Should be tomorrow at midnight
	expected := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	if !resetTime.Equal(expected) {
		t.Errorf("daily reset time incorrect: got %v, expected %v", resetTime, expected)
	}

	// Should be in the future
	if !resetTime.After(now) {
		t.Error("daily reset time should be in the future")
	}
}

func TestCalculateResetTimeWeekly(t *testing.T) {
	cfg := &config.Config{}
	agg := New(cfg, nil)

	now := time.Now()
	resetTime := agg.CalculateResetTime("weekly")

	// Should be next Monday at midnight
	if !resetTime.After(now) {
		t.Error("weekly reset time should be in the future")
	}

	// Should be a Monday
	if resetTime.Weekday() != time.Monday {
		t.Errorf("weekly reset should be on Monday, got %v", resetTime.Weekday())
	}

	// Should be at midnight
	if resetTime.Hour() != 0 || resetTime.Minute() != 0 || resetTime.Second() != 0 {
		t.Error("weekly reset should be at midnight")
	}

	t.Logf("Next weekly reset: %v", resetTime)
}

func TestCalculateResetTimeMonthly(t *testing.T) {
	cfg := &config.Config{}
	agg := New(cfg, nil)

	now := time.Now()
	resetTime := agg.CalculateResetTime("monthly")

	// Should be the 1st of next month at midnight
	if !resetTime.After(now) {
		t.Error("monthly reset time should be in the future")
	}

	// Should be the 1st day of the month
	if resetTime.Day() != 1 {
		t.Errorf("monthly reset should be on the 1st, got day %d", resetTime.Day())
	}

	// Should be at midnight
	if resetTime.Hour() != 0 || resetTime.Minute() != 0 || resetTime.Second() != 0 {
		t.Error("monthly reset should be at midnight")
	}

	t.Logf("Next monthly reset: %v", resetTime)
}
