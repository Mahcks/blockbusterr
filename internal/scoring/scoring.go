package scoring

import (
	"math"
	"time"

	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/integrations"
)

// ContentScore represents a scored item with its source job
type ContentScore struct {
	MediaType string // "movie" or "show"
	JobName   string // Source job name (trending_movies, popular_shows, etc.)
	Title     string
	Year      int
	TMDBID    int
	TVDBID    int
	IMDBID    string
	Score     float64
	Rank      int // Will be set after sorting

	// Original data for adding to Radarr/Sonarr
	MovieData *integrations.Movie
	ShowData  *integrations.Show
}

// Weights represents scoring weights configuration
type Weights struct {
	Rating     float64
	Popularity float64
	Recency    float64
}

// GetDefaultWeights returns the default scoring weights
func GetDefaultWeights() Weights {
	return Weights{
		Rating:     0.6,
		Popularity: 0.3,
		Recency:    0.1,
	}
}

// GetWeightsFromConfig returns scoring weights from config, or defaults if not configured
func GetWeightsFromConfig(cfg *config.Config) Weights {
	if !cfg.Scoring.Enabled {
		return GetDefaultWeights()
	}

	// Validate weights sum to ~1.0 (allow small tolerance)
	total := cfg.Scoring.RatingWeight + cfg.Scoring.PopularityWeight + cfg.Scoring.RecencyWeight
	if total < 0.99 || total > 1.01 {
		// Invalid weights, use defaults
		return GetDefaultWeights()
	}

	return Weights{
		Rating:     cfg.Scoring.RatingWeight,
		Popularity: cfg.Scoring.PopularityWeight,
		Recency:    cfg.Scoring.RecencyWeight,
	}
}

// CalculateMovieScore calculates a score for a movie based on configurable weights
// Returns a score between 0.0 and 1.0
func CalculateMovieScore(movie integrations.Movie, cfg *config.Config) float64 {
	weights := GetWeightsFromConfig(cfg)

	// Component 1: Rating score
	ratingScore := normalizeRating(movie.Rating, cfg)

	// Component 2: Popularity score
	popularityScore := normalizePopularity(movie.Votes, cfg)

	// Component 3: Recency score
	recencyScore := normalizeRecency(movie.Year, cfg)

	// Calculate final score
	finalScore := (ratingScore * weights.Rating) +
		(popularityScore * weights.Popularity) +
		(recencyScore * weights.Recency)

	// Ensure score is in 0-1 range
	if finalScore < 0 {
		finalScore = 0
	}
	if finalScore > 1 {
		finalScore = 1
	}

	return finalScore
}

// CalculateShowScore calculates a score for a TV show based on configurable weights
// Returns a score between 0.0 and 1.0
func CalculateShowScore(show integrations.Show, cfg *config.Config) float64 {
	weights := GetWeightsFromConfig(cfg)

	// Component 1: Rating score
	ratingScore := normalizeRating(show.Rating, cfg)

	// Component 2: Popularity score
	popularityScore := normalizePopularity(show.Votes, cfg)

	// Component 3: Recency score
	recencyScore := normalizeRecency(show.Year, cfg)

	// Calculate final score
	finalScore := (ratingScore * weights.Rating) +
		(popularityScore * weights.Popularity) +
		(recencyScore * weights.Recency)

	// Ensure score is in 0-1 range
	if finalScore < 0 {
		finalScore = 0
	}
	if finalScore > 1 {
		finalScore = 1
	}

	return finalScore
}

// normalizeRating normalizes a rating to 0-1 scale
func normalizeRating(rating float64, cfg *config.Config) float64 {
	if rating <= 0 {
		return 0
	}

	scale := 10.0
	if cfg.Scoring.Enabled && cfg.Scoring.RatingScale > 0 {
		scale = cfg.Scoring.RatingScale
	}

	normalized := rating / scale
	if normalized > 1 {
		normalized = 1
	}

	return normalized
}

// normalizePopularity normalizes popularity metrics to 0-1 scale using logarithmic scaling
func normalizePopularity(votes int, cfg *config.Config) float64 {
	if votes <= 0 {
		return 0
	}

	// Logarithmic scaling: 100 votes = 0.2, 1000 = 0.3, 10000 = 0.4, 100000 = 0.5, 1000000 = 0.6
	logVotes := math.Log10(float64(votes))
	normalized := logVotes / 10.0

	// Cap at 1.0
	if normalized > 1 {
		normalized = 1
	}

	return normalized
}

// normalizeRecency normalizes recency based on release year
func normalizeRecency(year int, cfg *config.Config) float64 {
	if year <= 0 {
		return 0
	}

	currentYear := time.Now().Year()
	yearsDiff := currentYear - year

	// If content is from the future, give it max recency score
	if yearsDiff < 0 {
		return 1.0
	}

	// Get recency window from config (default 365 days = 1 year)
	recencyDays := 365
	if cfg.Scoring.Enabled && cfg.Scoring.RecencyDays > 0 {
		recencyDays = cfg.Scoring.RecencyDays
	}

	recencyYears := float64(recencyDays) / 365.0

	// Linear decay: content within recency window gets proportional score
	if float64(yearsDiff) <= recencyYears {
		return 1.0 - (float64(yearsDiff) / recencyYears)
	}

	// Content older than recency window gets 0
	return 0
}
