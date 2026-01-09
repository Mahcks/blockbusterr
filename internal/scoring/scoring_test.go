package scoring

import (
	"testing"

	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/integrations"
)

func TestCalculateMovieScore(t *testing.T) {
	cfg := &config.Config{}

	// High quality movie with lots of votes
	highQuality := integrations.Movie{
		Title:    "Excellent Movie",
		Year:     2024,
		Rating:   9.0,
		Votes:    1000000,
		Language: "en",
		Genres:   []string{"drama"},
	}

	// Low quality movie with few votes
	lowQuality := integrations.Movie{
		Title:    "Poor Movie",
		Year:     2024,
		Rating:   4.0,
		Votes:    100,
		Language: "en",
	}

	highScore := CalculateMovieScore(highQuality, cfg)
	lowScore := CalculateMovieScore(lowQuality, cfg)

	// Basic checks
	if highScore < 0 || highScore > 1 {
		t.Errorf("high score out of range: %f", highScore)
	}

	if lowScore < 0 || lowScore > 1 {
		t.Errorf("low score out of range: %f", lowScore)
	}

	// High quality should score higher
	if highScore <= lowScore {
		t.Errorf("high quality movie should score higher: high=%f, low=%f", highScore, lowScore)
	}

	t.Logf("High quality: %.3f, Low quality: %.3f", highScore, lowScore)
}

func TestCalculateShowScore(t *testing.T) {
	cfg := &config.Config{}

	// High quality show
	highQuality := integrations.Show{
		Title:    "Great Show",
		Year:     2023,
		Rating:   9.0,
		Votes:    500000,
		Language: "en",
		Genres:   []string{"drama"},
	}

	// Low quality show
	lowQuality := integrations.Show{
		Title:    "Bad Show",
		Year:     2020,
		Rating:   4.0,
		Votes:    100,
		Language: "en",
	}

	highScore := CalculateShowScore(highQuality, cfg)
	lowScore := CalculateShowScore(lowQuality, cfg)

	// Basic checks
	if highScore < 0 || highScore > 1 {
		t.Errorf("high score out of range: %f", highScore)
	}

	if lowScore < 0 || lowScore > 1 {
		t.Errorf("low score out of range: %f", lowScore)
	}

	// High quality should score higher
	if highScore <= lowScore {
		t.Errorf("high quality show should score higher: high=%f, low=%f", highScore, lowScore)
	}

	t.Logf("High quality: %.3f, Low quality: %.3f", highScore, lowScore)
}

func TestScoreFactors(t *testing.T) {
	// Test that different factors affect the score
	cfg := &config.Config{}

	// Base movie (older than recency window)
	base := integrations.Movie{
		Title:  "Movie",
		Year:   2020,
		Rating: 7.0,
		Votes:  10000,
	}

	// Higher rating
	higherRating := base
	higherRating.Rating = 9.0

	// Recent release (current year - more recent)
	recent := base
	recent.Year = 2026 // Current year for the test

	// More popular
	morePopular := base
	morePopular.Votes = 100000

	baseScore := CalculateMovieScore(base, cfg)
	ratingScore := CalculateMovieScore(higherRating, cfg)
	recentScore := CalculateMovieScore(recent, cfg)
	popularScore := CalculateMovieScore(morePopular, cfg)

	// Higher rating should increase score
	if ratingScore <= baseScore {
		t.Errorf("Higher rating should increase score: base=%.3f, rating=%.3f", baseScore, ratingScore)
	}

	// Recent release should increase score
	if recentScore <= baseScore {
		t.Errorf("Recent release should increase score: base=%.3f, recent=%.3f", baseScore, recentScore)
	}

	// More votes should increase score
	if popularScore <= baseScore {
		t.Errorf("More votes should increase score: base=%.3f, popular=%.3f", baseScore, popularScore)
	}

	t.Logf("Base: %.3f, +HighRating: %.3f, +Recent: %.3f, +Popular: %.3f",
		baseScore, ratingScore, recentScore, popularScore)
}
