package jobs

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/filters"
	"github.com/mahcks/blockbusterr/internal/integrations"
)

func TestPreviewDeliveryReasonMatchesMode(t *testing.T) {
	result := filters.FilterResult{Passed: true}
	if got := previewDeliveryReason("direct", result); got != "Will add: Passed all configured rules" {
		t.Fatalf("direct reason = %q", got)
	}
	if got := previewDeliveryReason("jellyseerr", result); got != "Will request: Passed all configured rules" {
		t.Fatalf("Jellyseerr reason = %q", got)
	}
}

func TestPreviewMoviesReturnsRadarrLookupError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cfg := &config.Config{}
	cfg.Radarr.URL = server.URL
	response := PreviewResponse{}
	if err := previewMovies(t.Context(), cfg, nil, config.DynamicJob{Mode: "direct"}, nil, &response); err == nil {
		t.Fatal("expected Radarr lookup error")
	}
}

func TestHasTMDBConfigured(t *testing.T) {
	tests := []struct {
		name     string
		apiKey   string
		expected bool
	}{
		{
			name:     "TMDB configured",
			apiKey:   "test-api-key-123",
			expected: true,
		},
		{
			name:     "TMDB not configured",
			apiKey:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{}
			cfg.TMDB.APIKey = tt.apiKey

			result := hasTMDBConfigured(cfg)
			if result != tt.expected {
				t.Errorf("hasTMDBConfigured() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCreateMoviePreviewItem(t *testing.T) {
	cfg := &config.Config{}
	cfg.TMDB.APIKey = "" // No TMDB configured

	movie := integrations.Movie{
		Title: "Test Movie",
		Year:  2024,
		IDs: integrations.IDs{
			TMDB: 12345,
			IMDB: "tt1234567",
		},
		Genres:   []string{"Action", "Adventure"},
		Overview: "A test movie",
		Rating:   8.5,
		Votes:    1000,
		Runtime:  120,
	}

	item := createMoviePreviewItem(cfg, movie, 50000)

	if item.Title != "Test Movie" {
		t.Errorf("Title = %s, want Test Movie", item.Title)
	}
	if item.Year != 2024 {
		t.Errorf("Year = %d, want 2024", item.Year)
	}
	if item.TMDBID != 12345 {
		t.Errorf("TMDBID = %d, want 12345", item.TMDBID)
	}
	if item.IMDBID != "tt1234567" {
		t.Errorf("IMDBID = %s, want tt1234567", item.IMDBID)
	}
	if item.Rating != 8.5 {
		t.Errorf("Rating = %f, want 8.5", item.Rating)
	}
	if item.Popularity != 50000 {
		t.Errorf("Popularity = %d, want 50000", item.Popularity)
	}
	if len(item.Genres) != 2 {
		t.Errorf("Genres length = %d, want 2", len(item.Genres))
	}
	if item.PosterURL != "" {
		t.Errorf("PosterURL = %s, want empty string (no TMDB configured)", item.PosterURL)
	}
}

func TestCreateShowPreviewItem(t *testing.T) {
	cfg := &config.Config{}
	cfg.TMDB.APIKey = "" // No TMDB configured

	show := integrations.Show{
		Title: "Test Show",
		Year:  2024,
		IDs: integrations.IDs{
			TMDB: 12345,
			TVDB: 67890,
			IMDB: "tt1234567",
		},
		Genres:   []string{"Drama", "Thriller"},
		Overview: "A test show",
		Rating:   9.0,
		Votes:    5000,
		Runtime:  45,
	}

	item := createShowPreviewItem(cfg, show, 30000)

	if item.Title != "Test Show" {
		t.Errorf("Title = %s, want Test Show", item.Title)
	}
	if item.Year != 2024 {
		t.Errorf("Year = %d, want 2024", item.Year)
	}
	if item.TMDBID != 12345 {
		t.Errorf("TMDBID = %d, want 12345", item.TMDBID)
	}
	if item.TVDBID != 67890 {
		t.Errorf("TVDBID = %d, want 67890", item.TVDBID)
	}
	if item.IMDBID != "tt1234567" {
		t.Errorf("IMDBID = %s, want tt1234567", item.IMDBID)
	}
	if item.Rating != 9.0 {
		t.Errorf("Rating = %f, want 9.0", item.Rating)
	}
	if item.Popularity != 30000 {
		t.Errorf("Popularity = %d, want 30000", item.Popularity)
	}
	if len(item.Genres) != 2 {
		t.Errorf("Genres length = %d, want 2", len(item.Genres))
	}
}

func TestPreviewResponse(t *testing.T) {
	response := PreviewResponse{
		JobName:       "trending_movies",
		TotalFound:    25,
		WillAdd:       10,
		AlreadyExists: 8,
		FilteredOut:   7,
		Mode:          "direct",
		HasPosters:    true,
		Items:         []PreviewItem{},
	}

	if response.JobName != "trending_movies" {
		t.Errorf("JobName = %s, want trending_movies", response.JobName)
	}
	if response.TotalFound != 25 {
		t.Errorf("TotalFound = %d, want 25", response.TotalFound)
	}
	if response.WillAdd != 10 {
		t.Errorf("WillAdd = %d, want 10", response.WillAdd)
	}
	if response.AlreadyExists != 8 {
		t.Errorf("AlreadyExists = %d, want 8", response.AlreadyExists)
	}
	if response.FilteredOut != 7 {
		t.Errorf("FilteredOut = %d, want 7", response.FilteredOut)
	}
	if response.Mode != "direct" {
		t.Errorf("Mode = %s, want direct", response.Mode)
	}
	if !response.HasPosters {
		t.Errorf("HasPosters = false, want true")
	}

	// Verify counts add up
	total := response.WillAdd + response.AlreadyExists + response.FilteredOut
	if total != response.TotalFound {
		t.Errorf("Total counts = %d, want %d", total, response.TotalFound)
	}
}

func TestPreviewItem(t *testing.T) {
	item := PreviewItem{
		Title:         "Test Movie",
		Year:          2024,
		TMDBID:        12345,
		IMDBID:        "tt1234567",
		PosterURL:     "https://image.tmdb.org/...",
		Overview:      "Test overview",
		Rating:        8.5,
		Votes:         1000,
		Popularity:    50000,
		Genres:        []string{"Action"},
		Runtime:       120,
		AlreadyExists: false,
		FilteredOut:   false,
		FilterReason:  "",
	}

	if item.AlreadyExists {
		t.Error("AlreadyExists should be false")
	}
	if item.FilteredOut {
		t.Error("FilteredOut should be false")
	}
	if item.FilterReason != "" {
		t.Errorf("FilterReason should be empty, got %s", item.FilterReason)
	}
}

func TestPreviewItemFiltered(t *testing.T) {
	item := PreviewItem{
		Title:        "Bad Movie",
		Year:         2024,
		Rating:       3.0,
		FilteredOut:  true,
		FilterReason: "Rating 3.0 below minimum threshold",
	}

	if !item.FilteredOut {
		t.Error("FilteredOut should be true")
	}
	if item.FilterReason == "" {
		t.Error("FilterReason should not be empty for filtered items")
	}
	if item.FilterReason != "Rating 3.0 below minimum threshold" {
		t.Errorf("FilterReason = %s, want 'Rating 3.0 below minimum threshold'", item.FilterReason)
	}
}

func TestPreviewExplainsCertificationDecision(t *testing.T) {
	cfg := &config.Config{}
	cfg.Filters.Movies = config.MovieFilters{CertificationCountry: "US", AllowedCertifications: []string{"PG"}, UnknownCertification: "reject"}
	response := PreviewResponse{}
	movies := []integrations.Movie{{Title: "Rated", IDs: integrations.IDs{TMDB: 1}, Certifications: []integrations.Certification{{Value: "R", Country: "US", Source: "tmdb"}}}}
	if err := previewMovies(t.Context(), cfg, nil, config.DynamicJob{}, movies, &response); err != nil {
		t.Fatal(err)
	}
	if len(response.Items) != 1 || !response.Items[0].FilteredOut || len(response.Items[0].FilterChecks) == 0 || response.Items[0].FilterChecks[0].Message != "US certification R is not allowed (TMDB)" {
		t.Fatalf("preview = %#v", response)
	}
}
