package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/integrations"
	"github.com/mahcks/blockbusterr/internal/scoring"
)

func main() {
	fmt.Println("=== Content Scoring System Demo (LIVE DATA) ===")

	// Set to dev mode to load config.dev.yaml
	os.Setenv("VERSION", "dev")

	// Load actual configuration
	cfg, err := config.New("dev")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Configuration:\n")
	fmt.Printf("  Rating Weight:     %.1f (%.0f%%)\n", cfg.Scoring.RatingWeight, cfg.Scoring.RatingWeight*100)
	fmt.Printf("  Popularity Weight: %.1f (%.0f%%)\n", cfg.Scoring.PopularityWeight, cfg.Scoring.PopularityWeight*100)
	fmt.Printf("  Recency Weight:    %.1f (%.0f%%)\n", cfg.Scoring.RecencyWeight, cfg.Scoring.RecencyWeight*100)
	fmt.Printf("  Recency Window:    %d days\n\n", cfg.Scoring.RecencyDays)

	// Initialize Trakt client
	fmt.Println("🔍 Fetching live data from Trakt...")
	traktClient := integrations.NewTrakt(integrations.TraktConfig{
		ClientID:     cfg.Trakt.ClientID,
		ClientSecret: cfg.Trakt.ClientSecret,
	})

	ctx := context.Background()

	// Fetch trending movies
	traktMovies, err := traktClient.GetTrendingMovies(ctx, 20)
	if err != nil {
		log.Fatalf("Failed to fetch trending movies: %v", err)
	}

	fmt.Printf("✅ Fetched %d trending movies from Trakt\n\n", len(traktMovies))

	// Extract the actual movies from TrendingMovie wrapper
	var movies []integrations.Movie
	for _, tm := range traktMovies {
		if len(movies) >= 15 { // Limit to 15 for display
			break
		}
		movies = append(movies, tm.Movie)
	}

	fmt.Printf("📊 Scoring %d movies...\n\n", len(movies))
	var scores []scoring.ContentScore
	for _, movie := range movies {
		score := scoring.CalculateMovieScore(movie, cfg)
		scores = append(scores, scoring.ContentScore{
			MediaType: "movie",
			JobName:   "trending_movies",
			Title:     movie.Title,
			Year:      movie.Year,
			Score:     score,
			MovieData: &movie,
		})
	}

	// Sort by score descending
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})

	// Assign ranks
	for i := range scores {
		scores[i].Rank = i + 1
	}

	// Display results
	fmt.Println("=== Scored Results (Live Trending Movies) ===")
	fmt.Printf("%-4s %-45s %-6s %-10s %-12s %-8s\n", "Rank", "Title", "Year", "Rating", "Votes", "Score")
	fmt.Println("──────────────────────────────────────────────────────────────────────────────────────────")

	for _, item := range scores {
		movie := item.MovieData
		scorePercent := item.Score * 100

		// Color code the score
		var scoreColor string
		if item.Score >= 0.8 {
			scoreColor = "🟢" // Excellent
		} else if item.Score >= 0.6 {
			scoreColor = "🔵" // Good
		} else if item.Score >= 0.4 {
			scoreColor = "🟡" // Fair
		} else {
			scoreColor = "⚪" // Poor
		}

		fmt.Printf("#%-3d %-45s %-6d %-10.1f %-12s %s %.1f%%\n",
			item.Rank,
			truncate(item.Title, 45),
			item.Year,
			movie.Rating,
			formatVotes(movie.Votes),
			scoreColor,
			scorePercent,
		)
	}

	// Test different presets
	fmt.Println("\n\n=== Testing Different Presets ===")

	presets := []struct {
		name    string
		rating  float64
		popular float64
		recency float64
	}{
		{"Quality-Focused", 0.8, 0.1, 0.1},
		{"Trending-Focused", 0.3, 0.6, 0.1},
		{"New Releases", 0.4, 0.2, 0.4},
	}

	// Save original weights
	origRating := cfg.Scoring.RatingWeight
	origPopular := cfg.Scoring.PopularityWeight
	origRecency := cfg.Scoring.RecencyWeight

	for _, preset := range presets {
		cfg.Scoring.RatingWeight = preset.rating
		cfg.Scoring.PopularityWeight = preset.popular
		cfg.Scoring.RecencyWeight = preset.recency

		// Recalculate scores with new weights
		var newScores []scoring.ContentScore
		for _, movie := range movies {
			score := scoring.CalculateMovieScore(movie, cfg)
			newScores = append(newScores, scoring.ContentScore{
				Title:     movie.Title,
				Year:      movie.Year,
				Score:     score,
				MovieData: &movie,
			})
		}

		// Sort and rank
		sort.Slice(newScores, func(i, j int) bool {
			return newScores[i].Score > newScores[j].Score
		})
		for i := range newScores {
			newScores[i].Rank = i + 1
		}

		// Display top 5
		fmt.Printf("\n%s (R:%.1f P:%.1f N:%.1f):\n", preset.name, preset.rating, preset.popular, preset.recency)
		for i := 0; i < 5 && i < len(newScores); i++ {
			item := newScores[i]
			movie := item.MovieData
			fmt.Printf("  #%d  %s (%.1f%%, %d, ⭐%.1f)\n",
				item.Rank,
				truncate(item.Title, 35),
				item.Score*100,
				item.Year,
				movie.Rating,
			)
		}
	}

	// Restore original weights
	cfg.Scoring.RatingWeight = origRating
	cfg.Scoring.PopularityWeight = origPopular
	cfg.Scoring.RecencyWeight = origRecency

	fmt.Println("\n=== Summary ===")
	fmt.Printf("✅ Successfully scored %d live movies from Trakt\n", len(movies))
	fmt.Println("💡 Different presets show how weight changes affect rankings")
	fmt.Println("📈 This is the same algorithm used by scheduled jobs")
}

func truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length-3] + "..."
}

func formatVotes(votes int) string {
	if votes >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(votes)/1000000)
	}
	if votes >= 1000 {
		return fmt.Sprintf("%.0fK", float64(votes)/1000)
	}
	return fmt.Sprintf("%d", votes)
}
