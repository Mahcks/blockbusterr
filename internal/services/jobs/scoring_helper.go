package jobs

import (
	"context"
	"sort"

	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/integrations"
	"github.com/mahcks/blockbusterr/internal/scoring"
)

// ScoreAndRankMovies scores and ranks a list of movies if scoring is enabled in config
// Returns a map of TMDB ID -> (score, rank) for easy lookup during processing
func ScoreAndRankMovies(movies []integrations.Movie, cfg *config.Config) map[int]ScoreInfo {
	result := make(map[int]ScoreInfo)

	// If scoring is disabled, return empty map
	if !cfg.Scoring.Enabled {
		return result
	}

	// Calculate scores for all movies
	var scoredItems []scoring.ContentScore
	for _, movie := range movies {
		score := scoring.CalculateMovieScore(movie, cfg)
		scoredItems = append(scoredItems, scoring.ContentScore{
			Title:     movie.Title,
			Year:      movie.Year,
			Score:     score,
			MovieData: &movie,
		})
	}

	// Sort by score descending
	sort.Slice(scoredItems, func(i, j int) bool {
		return scoredItems[i].Score > scoredItems[j].Score
	})

	// Assign ranks and store in result map
	for i, item := range scoredItems {
		if item.MovieData != nil && item.MovieData.IDs.TMDB > 0 {
			result[item.MovieData.IDs.TMDB] = ScoreInfo{
				Score: item.Score,
				Rank:  i + 1,
			}
		}
	}

	return result
}

// ScoreAndRankShows scores and ranks a list of TV shows if scoring is enabled in config
// Returns a map of TVDB ID -> (score, rank) for easy lookup during processing
func ScoreAndRankShows(shows []integrations.Show, cfg *config.Config) map[int]ScoreInfo {
	result := make(map[int]ScoreInfo)

	// If scoring is disabled, return empty map
	if !cfg.Scoring.Enabled {
		return result
	}

	// Calculate scores for all shows
	var scoredItems []scoring.ContentScore
	for _, show := range shows {
		score := scoring.CalculateShowScore(show, cfg)
		scoredItems = append(scoredItems, scoring.ContentScore{
			Title:    show.Title,
			Year:     show.Year,
			Score:    score,
			ShowData: &show,
		})
	}

	// Sort by score descending
	sort.Slice(scoredItems, func(i, j int) bool {
		return scoredItems[i].Score > scoredItems[j].Score
	})

	// Assign ranks and store in result map
	for i, item := range scoredItems {
		if item.ShowData != nil && item.ShowData.IDs.TVDB > 0 {
			result[item.ShowData.IDs.TVDB] = ScoreInfo{
				Score: item.Score,
				Rank:  i + 1,
			}
		}
	}

	return result
}

// GetTMDBPosterURL fetches the actual poster URL from TMDB API for activity logs
func GetTMDBPosterURL(cfg *config.Config, tmdbID int, mediaType string) string {
	if cfg.TMDB.APIKey == "" || tmdbID == 0 {
		return ""
	}

	tmdbClient := integrations.NewTMDB(integrations.TMDBConfig{
		APIKey: cfg.TMDB.APIKey,
	})

	ctx := context.Background()
	var posterURL string
	var err error

	if mediaType == "movie" {
		posterURL, err = tmdbClient.GetMoviePosterURL(ctx, tmdbID)
	} else {
		posterURL, err = tmdbClient.GetShowPosterURL(ctx, tmdbID)
	}

	if err != nil {
		return ""
	}

	return posterURL
}

// GetShowPosterURL gets poster for show - only uses TMDB since TVDB images require authentication
func GetShowPosterURL(cfg *config.Config, tmdbID int, tvdbID int) string {
	// Only use TMDB (TVDB artworks require authentication)
	if tmdbID > 0 && cfg.TMDB.APIKey != "" {
		return GetTMDBPosterURL(cfg, tmdbID, "show")
	}

	return ""
}
