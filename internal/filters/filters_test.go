package filters

import (
	"testing"

	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/integrations"
)

func TestMoviePassesFilters_AllowedCountries(t *testing.T) {
	tests := []struct {
		name             string
		movieCountry     string
		allowedCountries []string
		shouldPass       bool
	}{
		{
			name:             "US movie with US allowed",
			movieCountry:     "us",
			allowedCountries: []string{"us", "gb"},
			shouldPass:       true,
		},
		{
			name:             "FR movie with US allowed only",
			movieCountry:     "fr",
			allowedCountries: []string{"us"},
			shouldPass:       false,
		},
		{
			name:             "No country restrictions",
			movieCountry:     "jp",
			allowedCountries: []string{},
			shouldPass:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			movie := integrations.Movie{
				Title:   "Test Movie",
				Country: tt.movieCountry,
			}

			filters := config.MovieFilters{
				AllowedCountries: tt.allowedCountries,
			}

			passes, _ := MoviePassesFilters(movie, filters)
			if passes != tt.shouldPass {
				t.Errorf("MoviePassesFilters() = %v, want %v", passes, tt.shouldPass)
			}
		})
	}
}

func TestMoviePassesFilters_BlacklistedGenres(t *testing.T) {
	tests := []struct {
		name              string
		movieGenres       []string
		blacklistedGenres []string
		shouldPass        bool
	}{
		{
			name:              "Action movie with Documentary blacklisted",
			movieGenres:       []string{"action", "thriller"},
			blacklistedGenres: []string{"documentary"},
			shouldPass:        true,
		},
		{
			name:              "Documentary with Documentary blacklisted",
			movieGenres:       []string{"documentary"},
			blacklistedGenres: []string{"documentary"},
			shouldPass:        false,
		},
		{
			name:              "Multiple genres with one blacklisted",
			movieGenres:       []string{"action", "documentary"},
			blacklistedGenres: []string{"documentary"},
			shouldPass:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			movie := integrations.Movie{
				Title:  "Test Movie",
				Genres: tt.movieGenres,
			}

			filters := config.MovieFilters{
				BlacklistedGenres: tt.blacklistedGenres,
			}

			passes, _ := MoviePassesFilters(movie, filters)
			if passes != tt.shouldPass {
				t.Errorf("MoviePassesFilters() = %v, want %v", passes, tt.shouldPass)
			}
		})
	}
}

func TestMoviePassesFilters_Runtime(t *testing.T) {
	tests := []struct {
		name       string
		runtime    int
		minRuntime int
		maxRuntime int
		shouldPass bool
	}{
		{
			name:       "90 min movie with 80-120 range",
			runtime:    90,
			minRuntime: 80,
			maxRuntime: 120,
			shouldPass: true,
		},
		{
			name:       "60 min movie with 80-120 range",
			runtime:    60,
			minRuntime: 80,
			maxRuntime: 120,
			shouldPass: false,
		},
		{
			name:       "150 min movie with 80-120 range",
			runtime:    150,
			minRuntime: 80,
			maxRuntime: 120,
			shouldPass: false,
		},
		{
			name:       "90 min movie with no restrictions",
			runtime:    90,
			minRuntime: 0,
			maxRuntime: 0,
			shouldPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			movie := integrations.Movie{
				Title:   "Test Movie",
				Runtime: tt.runtime,
			}

			filters := config.MovieFilters{
				BlacklistedMinRuntime: tt.minRuntime,
				BlacklistedMaxRuntime: tt.maxRuntime,
			}

			passes, _ := MoviePassesFilters(movie, filters)
			if passes != tt.shouldPass {
				t.Errorf("MoviePassesFilters() = %v, want %v", passes, tt.shouldPass)
			}
		})
	}
}

func TestMoviePassesFilters_Year(t *testing.T) {
	tests := []struct {
		name       string
		year       int
		minYear    int
		maxYear    int
		shouldPass bool
	}{
		{
			name:       "2024 movie with 2020-2025 range",
			year:       2024,
			minYear:    2020,
			maxYear:    2025,
			shouldPass: true,
		},
		{
			name:       "2019 movie with 2020-2025 range",
			year:       2019,
			minYear:    2020,
			maxYear:    2025,
			shouldPass: false,
		},
		{
			name:       "2026 movie with 2020-2025 range",
			year:       2026,
			minYear:    2020,
			maxYear:    2025,
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			movie := integrations.Movie{
				Title: "Test Movie",
				Year:  tt.year,
			}

			filters := config.MovieFilters{
				BlacklistedMinYear: tt.minYear,
				BlacklistedMaxYear: tt.maxYear,
			}

			passes, _ := MoviePassesFilters(movie, filters)
			if passes != tt.shouldPass {
				t.Errorf("MoviePassesFilters() = %v, want %v", passes, tt.shouldPass)
			}
		})
	}
}

func TestShowPassesFilters_BlacklistedNetworks(t *testing.T) {
	tests := []struct {
		name                string
		network             string
		blacklistedNetworks []string
		shouldPass          bool
	}{
		{
			name:                "HBO show with Hallmark blacklisted",
			network:             "HBO",
			blacklistedNetworks: []string{"Hallmark", "Nickelodeon"},
			shouldPass:          true,
		},
		{
			name:                "Hallmark show with Hallmark blacklisted",
			network:             "Hallmark",
			blacklistedNetworks: []string{"Hallmark"},
			shouldPass:          false,
		},
		{
			name:                "Case insensitive matching",
			network:             "hallmark",
			blacklistedNetworks: []string{"Hallmark"},
			shouldPass:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			show := integrations.Show{
				Title:   "Test Show",
				Network: tt.network,
			}

			filters := config.ShowFilters{
				BlacklistedNetworks: tt.blacklistedNetworks,
			}

			passes, _ := ShowPassesFilters(show, filters)
			if passes != tt.shouldPass {
				t.Errorf("ShowPassesFilters() = %v, want %v", passes, tt.shouldPass)
			}
		})
	}
}

func TestMoviePassesFilters_Keywords(t *testing.T) {
	tests := []struct {
		name                string
		title               string
		blacklistedKeywords []string
		shouldPass          bool
	}{
		{
			name:                "Avengers with Marvel blacklisted",
			title:               "Avengers: Endgame",
			blacklistedKeywords: []string{"marvel"},
			shouldPass:          true, // "marvel" not in title
		},
		{
			name:                "Christmas Movie with christmas blacklisted",
			title:               "A Christmas Story",
			blacklistedKeywords: []string{"christmas"},
			shouldPass:          false,
		},
		{
			name:                "Case insensitive keyword matching",
			title:               "CHRISTMAS SPECIAL",
			blacklistedKeywords: []string{"christmas"},
			shouldPass:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			movie := integrations.Movie{
				Title: tt.title,
			}

			filters := config.MovieFilters{
				BlacklistedKeywords: tt.blacklistedKeywords,
			}

			passes, _ := MoviePassesFilters(movie, filters)
			if passes != tt.shouldPass {
				t.Errorf("MoviePassesFilters() = %v, want %v", passes, tt.shouldPass)
			}
		})
	}
}

func TestMoviePassesFilters_TMDBID(t *testing.T) {
	tests := []struct {
		name           string
		tmdbID         int
		blacklistedIDs []int
		shouldPass     bool
	}{
		{
			name:           "Movie with blacklisted ID",
			tmdbID:         12345,
			blacklistedIDs: []int{12345, 67890},
			shouldPass:     false,
		},
		{
			name:           "Movie with non-blacklisted ID",
			tmdbID:         99999,
			blacklistedIDs: []int{12345, 67890},
			shouldPass:     true,
		},
		{
			name:           "No blacklisted IDs",
			tmdbID:         12345,
			blacklistedIDs: []int{},
			shouldPass:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			movie := integrations.Movie{
				Title: "Test Movie",
				IDs: integrations.IDs{
					TMDB: tt.tmdbID,
				},
			}

			filters := config.MovieFilters{
				BlacklistedTMDBIds: tt.blacklistedIDs,
			}

			passes, _ := MoviePassesFilters(movie, filters)
			if passes != tt.shouldPass {
				t.Errorf("MoviePassesFilters() = %v, want %v", passes, tt.shouldPass)
			}
		})
	}
}

func TestMoviePassesFilters_MultipleFilters(t *testing.T) {
	// Test that all filters must pass
	movie := integrations.Movie{
		Title:   "Test Documentary",
		Year:    2024,
		Country: "us",
		Genres:  []string{"documentary"},
		Runtime: 90,
	}

	filters := config.MovieFilters{
		AllowedCountries:   []string{"us"},          // PASS
		BlacklistedGenres:  []string{"documentary"}, // FAIL
		BlacklistedMinYear: 2020,                    // PASS
		BlacklistedMaxYear: 2025,                    // PASS
	}

	passes, reason := MoviePassesFilters(movie, filters)
	if passes {
		t.Error("Movie should not pass when one filter fails")
	}
	if reason == "" {
		t.Error("Filter reason should be provided")
	}
}

func TestMovieRatingFilter(t *testing.T) {
	tests := []struct {
		name     string
		movie    integrations.Movie
		filter   config.MovieFilters
		expected bool
		reason   string
	}{
		{
			name: "Movie above rating threshold",
			movie: integrations.Movie{
				Title:  "Great Movie",
				Year:   2024,
				Rating: 8.5,
				Votes:  5000,
			},
			filter: config.MovieFilters{
				MinRating: 7.0,
				MinVotes:  1000,
			},
			expected: true,
			reason:   "",
		},
		{
			name: "Movie below rating threshold",
			movie: integrations.Movie{
				Title:  "Bad Movie",
				Year:   2024,
				Rating: 5.2,
				Votes:  5000,
			},
			filter: config.MovieFilters{
				MinRating: 6.5,
				MinVotes:  1000,
			},
			expected: false,
			reason:   "rating 5.2 below minimum 6.5",
		},
		{
			name: "Movie with insufficient votes",
			movie: integrations.Movie{
				Title:  "Obscure Movie",
				Year:   2024,
				Rating: 8.0,
				Votes:  500,
			},
			filter: config.MovieFilters{
				MinRating: 6.5,
				MinVotes:  1000,
			},
			expected: false,
			reason:   "votes 500 below minimum 1000",
		},
		{
			name: "Rating filter disabled (0)",
			movie: integrations.Movie{
				Title:  "Any Movie",
				Year:   2024,
				Rating: 3.0,
				Votes:  100,
			},
			filter: config.MovieFilters{
				MinRating: 0,
				MinVotes:  0,
			},
			expected: true,
			reason:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			passes, reason := MoviePassesFilters(tt.movie, tt.filter)
			if passes != tt.expected {
				t.Errorf("MoviePassesFilters() passes = %v, expected %v", passes, tt.expected)
			}
			if passes == false && reason != tt.reason {
				t.Errorf("MoviePassesFilters() reason = %v, expected %v", reason, tt.reason)
			}
		})
	}
}

func TestShowRatingFilter(t *testing.T) {
	tests := []struct {
		name     string
		show     integrations.Show
		filter   config.ShowFilters
		expected bool
		reason   string
	}{
		{
			name: "Show above rating threshold",
			show: integrations.Show{
				Title:  "Great Show",
				Year:   2024,
				Rating: 8.5,
				Votes:  3000,
			},
			filter: config.ShowFilters{
				MinRating: 7.0,
				MinVotes:  500,
			},
			expected: true,
			reason:   "",
		},
		{
			name: "Show below rating threshold",
			show: integrations.Show{
				Title:  "Bad Show",
				Year:   2024,
				Rating: 6.0,
				Votes:  3000,
			},
			filter: config.ShowFilters{
				MinRating: 7.0,
				MinVotes:  500,
			},
			expected: false,
			reason:   "rating 6.0 below minimum 7.0",
		},
		{
			name: "Show with insufficient votes",
			show: integrations.Show{
				Title:  "Niche Show",
				Year:   2024,
				Rating: 8.5,
				Votes:  200,
			},
			filter: config.ShowFilters{
				MinRating: 7.0,
				MinVotes:  500,
			},
			expected: false,
			reason:   "votes 200 below minimum 500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			passes, reason := ShowPassesFilters(tt.show, tt.filter)
			if passes != tt.expected {
				t.Errorf("ShowPassesFilters() passes = %v, expected %v", passes, tt.expected)
			}
			if passes == false && reason != tt.reason {
				t.Errorf("ShowPassesFilters() reason = %v, expected %v", reason, tt.reason)
			}
		})
	}
}

func TestFilterMoviesWithRatings(t *testing.T) {
	movies := []integrations.Movie{
		{Title: "Great Movie", Year: 2024, Rating: 8.5, Votes: 5000},
		{Title: "Good Movie", Year: 2024, Rating: 7.2, Votes: 3000},
		{Title: "Bad Movie", Year: 2024, Rating: 5.0, Votes: 2000},
		{Title: "Obscure Movie", Year: 2024, Rating: 8.0, Votes: 100},
	}

	filters := config.MovieFilters{
		MinRating: 7.0,
		MinVotes:  1000,
	}

	filtered := FilterMovies(movies, filters)

	if len(filtered) != 2 {
		t.Errorf("Expected 2 movies to pass filters, got %d", len(filtered))
	}

	// Check that only the correct movies passed
	expectedTitles := map[string]bool{
		"Great Movie": true,
		"Good Movie":  true,
	}

	for _, movie := range filtered {
		if !expectedTitles[movie.Title] {
			t.Errorf("Unexpected movie in filtered list: %s", movie.Title)
		}
	}
}
