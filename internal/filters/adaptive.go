package filters

import (
	"fmt"
	"math"
	"sort"

	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/integrations"
)

// PopularityPercentile represents the popularity ranking of content
type PopularityPercentile struct {
	TMDBID     int     // For movies
	TVDBID     int     // For shows
	Percentile float64 // 0.0 to 1.0 (0 = least popular, 1 = most popular)
}

// CalculateMoviePopularityPercentiles calculates popularity percentiles for a set of movies
// based on their vote counts. Returns a map of TMDB ID to percentile (0.0-1.0)
func CalculateMoviePopularityPercentiles(movies []integrations.Movie) map[int]float64 {
	if len(movies) == 0 {
		return make(map[int]float64)
	}

	// Create a slice of vote counts with their TMDB IDs
	type movieVotes struct {
		TMDBID int
		Votes  int
	}

	votesData := make([]movieVotes, len(movies))
	for i, movie := range movies {
		votesData[i] = movieVotes{
			TMDBID: movie.IDs.TMDB,
			Votes:  movie.Votes,
		}
	}

	// Sort by votes (ascending)
	sort.Slice(votesData, func(i, j int) bool {
		return votesData[i].Votes < votesData[j].Votes
	})

	// Calculate percentile for each movie
	percentiles := make(map[int]float64)
	totalMovies := float64(len(votesData))

	for rank, data := range votesData {
		// Percentile = rank / (total - 1)
		// rank 0 (lowest votes) = 0.0 percentile
		// rank n-1 (highest votes) = 1.0 percentile
		var percentile float64
		if totalMovies > 1 {
			percentile = float64(rank) / (totalMovies - 1)
		} else {
			percentile = 0.5 // Single item defaults to middle
		}

		percentiles[data.TMDBID] = percentile
	}

	return percentiles
}

// CalculateShowPopularityPercentiles calculates popularity percentiles for a set of shows
// based on their vote counts. Returns a map of TVDB ID to percentile (0.0-1.0)
func CalculateShowPopularityPercentiles(shows []integrations.Show) map[int]float64 {
	if len(shows) == 0 {
		return make(map[int]float64)
	}

	// Create a slice of vote counts with their TVDB IDs
	type showVotes struct {
		TVDBID int
		Votes  int
	}

	votesData := make([]showVotes, len(shows))
	for i, show := range shows {
		votesData[i] = showVotes{
			TVDBID: show.IDs.TVDB,
			Votes:  show.Votes,
		}
	}

	// Sort by votes (ascending)
	sort.Slice(votesData, func(i, j int) bool {
		return votesData[i].Votes < votesData[j].Votes
	})

	// Calculate percentile for each show
	percentiles := make(map[int]float64)
	totalShows := float64(len(votesData))

	for rank, data := range votesData {
		var percentile float64
		if totalShows > 1 {
			percentile = float64(rank) / (totalShows - 1)
		} else {
			percentile = 0.5
		}

		percentiles[data.TVDBID] = percentile
	}

	return percentiles
}

// CalculateAdaptiveRating calculates the adjusted rating threshold based on popularity
// Uses sliding scale: adjustment = (percentile - 0.5) * adjustmentFactor
//
// Examples with adjustmentFactor = 2.0:
// - Percentile 1.0 (most popular): baseRating - (0.5 * 2.0) = baseRating - 1.0
// - Percentile 0.5 (middle): baseRating - (0.0 * 2.0) = baseRating
// - Percentile 0.0 (least popular): baseRating - (-0.5 * 2.0) = baseRating + 1.0
func CalculateAdaptiveRating(baseMinRating, popularityPercentile, adjustmentFactor float64) float64 {
	// Calculate adjustment from the midpoint (0.5)
	adjustment := (popularityPercentile - 0.5) * adjustmentFactor

	// Apply adjustment (subtract because high popularity should lower threshold)
	adaptiveRating := baseMinRating - adjustment

	// Ensure rating doesn't go below 0 or above 10
	if adaptiveRating < 0 {
		adaptiveRating = 0
	}
	if adaptiveRating > 10 {
		adaptiveRating = 10
	}

	return math.Round(adaptiveRating*10) / 10 // Round to 1 decimal place
}

// MoviePassesAdaptiveFilters checks if a movie passes all configured filters with an adaptive rating threshold
func MoviePassesAdaptiveFilters(
	movie integrations.Movie,
	filters config.MovieFilters,
	adjustedMinRating float64,
) FilterResult {
	// Create a copy of filters with the adjusted rating
	adaptiveFilters := filters
	adaptiveFilters.MinRating = adjustedMinRating

	// Use the existing detailed filter logic
	result := MoviePassesFiltersDetailed(movie, adaptiveFilters)

	// Update the rating check message to show it was adaptive
	for i := range result.Checks {
		if result.Checks[i].Name == "Minimum Rating" {
			if result.Checks[i].Passed {
				result.Checks[i].Message = fmt.Sprintf(
					"Rating %.1f meets adaptive minimum %.1f",
					movie.Rating,
					adjustedMinRating,
				)
			} else {
				result.Checks[i].Message = fmt.Sprintf(
					"Rating %.1f below adaptive minimum %.1f",
					movie.Rating,
					adjustedMinRating,
				)
			}
			break
		}
	}

	return result
}

// ShowPassesAdaptiveFilters checks if a show passes all configured filters with an adaptive rating threshold
func ShowPassesAdaptiveFilters(
	show integrations.Show,
	filters config.ShowFilters,
	adjustedMinRating float64,
) FilterResult {
	// Create a copy of filters with the adjusted rating
	adaptiveFilters := filters
	adaptiveFilters.MinRating = adjustedMinRating

	// Use the existing detailed filter logic
	result := ShowPassesFiltersDetailed(show, adaptiveFilters)

	// Update the rating check message to show it was adaptive
	for i := range result.Checks {
		if result.Checks[i].Name == "Minimum Rating" {
			if result.Checks[i].Passed {
				result.Checks[i].Message = fmt.Sprintf(
					"Rating %.1f meets adaptive minimum %.1f",
					show.Rating,
					adjustedMinRating,
				)
			} else {
				result.Checks[i].Message = fmt.Sprintf(
					"Rating %.1f below adaptive minimum %.1f",
					show.Rating,
					adjustedMinRating,
				)
			}
			break
		}
	}

	return result
}
