package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// MovieFilters represents filtering options for movies
type MovieFilters struct {
	AllowedCountries      []string `mapstructure:"allowed_countries" json:"allowed_countries" yaml:"allowed_countries"`
	AllowedLanguages      []string `mapstructure:"allowed_languages" json:"allowed_languages" yaml:"allowed_languages"`
	BlacklistedGenres     []string `mapstructure:"blacklisted_genres" json:"blacklisted_genres" yaml:"blacklisted_genres"`
	BlacklistedKeywords   []string `mapstructure:"blacklisted_keywords" json:"blacklisted_keywords" yaml:"blacklisted_keywords"`
	BlacklistedTMDBIds    []int    `mapstructure:"blacklisted_tmdb_ids" json:"blacklisted_tmdb_ids" yaml:"blacklisted_tmdb_ids"`
	BlacklistedMinRuntime int      `mapstructure:"blacklisted_min_runtime" json:"blacklisted_min_runtime" yaml:"blacklisted_min_runtime"`
	BlacklistedMaxRuntime int      `mapstructure:"blacklisted_max_runtime" json:"blacklisted_max_runtime" yaml:"blacklisted_max_runtime"`
	BlacklistedMinYear    int      `mapstructure:"blacklisted_min_year" json:"blacklisted_min_year" yaml:"blacklisted_min_year"`
	BlacklistedMaxYear    int      `mapstructure:"blacklisted_max_year" json:"blacklisted_max_year" yaml:"blacklisted_max_year"`
	MinRating             float64  `mapstructure:"min_rating" json:"min_rating" yaml:"min_rating"`
	MinVotes              int      `mapstructure:"min_votes" json:"min_votes" yaml:"min_votes"`
}

// ShowFilters represents filtering options for TV shows
type ShowFilters struct {
	AllowedCountries      []string `mapstructure:"allowed_countries" json:"allowed_countries" yaml:"allowed_countries"`
	AllowedLanguages      []string `mapstructure:"allowed_languages" json:"allowed_languages" yaml:"allowed_languages"`
	BlacklistedGenres     []string `mapstructure:"blacklisted_genres" json:"blacklisted_genres" yaml:"blacklisted_genres"`
	BlacklistedKeywords   []string `mapstructure:"blacklisted_keywords" json:"blacklisted_keywords" yaml:"blacklisted_keywords"`
	BlacklistedNetworks   []string `mapstructure:"blacklisted_networks" json:"blacklisted_networks" yaml:"blacklisted_networks"`
	BlacklistedTVDBIds    []int    `mapstructure:"blacklisted_tvdb_ids" json:"blacklisted_tvdb_ids" yaml:"blacklisted_tvdb_ids"`
	BlacklistedMinRuntime int      `mapstructure:"blacklisted_min_runtime" json:"blacklisted_min_runtime" yaml:"blacklisted_min_runtime"`
	BlacklistedMaxRuntime int      `mapstructure:"blacklisted_max_runtime" json:"blacklisted_max_runtime" yaml:"blacklisted_max_runtime"`
	BlacklistedMinYear    int      `mapstructure:"blacklisted_min_year" json:"blacklisted_min_year" yaml:"blacklisted_min_year"`
	BlacklistedMaxYear    int      `mapstructure:"blacklisted_max_year" json:"blacklisted_max_year" yaml:"blacklisted_max_year"`
	MinRating             float64  `mapstructure:"min_rating" json:"min_rating" yaml:"min_rating"`
	MinVotes              int      `mapstructure:"min_votes" json:"min_votes" yaml:"min_votes"`
}

// Config represents the application configuration
type Config struct {
	Version string `mapstructure:"version" yaml:"version,omitempty"`

	Trakt struct {
		ClientID     string `mapstructure:"client_id" json:"client_id" yaml:"client_id"`
		ClientSecret string `mapstructure:"client_secret" json:"client_secret" yaml:"client_secret"`
	} `mapstructure:"trakt" json:"trakt" yaml:"trakt"`

	TMDB struct {
		APIKey string `mapstructure:"api_key" json:"api_key" yaml:"api_key"`
	} `mapstructure:"tmdb" json:"tmdb" yaml:"tmdb"`

	Radarr struct {
		URL                 string `mapstructure:"url" json:"url" yaml:"url"`
		APIKey              string `mapstructure:"api_key" json:"api_key" yaml:"api_key"`
		QualityProfile      int    `mapstructure:"quality_profile" json:"quality_profile" yaml:"quality_profile"`
		RootFolder          string `mapstructure:"root_folder" json:"root_folder" yaml:"root_folder"`
		MinimumAvailability string `mapstructure:"minimum_availability" json:"minimum_availability" yaml:"minimum_availability,omitempty"`
	} `mapstructure:"radarr" json:"radarr" yaml:"radarr"`

	Sonarr struct {
		URL            string `mapstructure:"url" json:"url" yaml:"url"`
		APIKey         string `mapstructure:"api_key" json:"api_key" yaml:"api_key"`
		QualityProfile int    `mapstructure:"quality_profile" json:"quality_profile" yaml:"quality_profile"`
		RootFolder     string `mapstructure:"root_folder" json:"root_folder" yaml:"root_folder"`
	} `mapstructure:"sonarr" json:"sonarr" yaml:"sonarr"`

	Jellyseerr struct {
		URL                string `mapstructure:"url" json:"url" yaml:"url"`
		APIKey             string `mapstructure:"api_key" json:"api_key" yaml:"api_key"`
		UserID             string `mapstructure:"user_id" json:"user_id" yaml:"user_id"` // Optional: request as specific user
		RequestCredentials struct {
			Email    string `mapstructure:"email" json:"email" yaml:"email"`
			Password string `mapstructure:"password" json:"password" yaml:"password"`
		} `mapstructure:"request_credentials" json:"request_credentials" yaml:"request_credentials,omitempty"`
	} `mapstructure:"jellyseerr" json:"jellyseerr" yaml:"jellyseerr"`

	Scoring struct {
		Enabled          bool    `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
		RatingWeight     float64 `mapstructure:"rating_weight" json:"rating_weight" yaml:"rating_weight"`
		PopularityWeight float64 `mapstructure:"popularity_weight" json:"popularity_weight" yaml:"popularity_weight"`
		RecencyWeight    float64 `mapstructure:"recency_weight" json:"recency_weight" yaml:"recency_weight"`
		RatingScale      float64 `mapstructure:"rating_scale" json:"rating_scale" yaml:"rating_scale"`
		PopularityMetric string  `mapstructure:"popularity_metric" json:"popularity_metric" yaml:"popularity_metric"` // votes, views, watchers
		RecencyDays      int     `mapstructure:"recency_days" json:"recency_days" yaml:"recency_days"`
	} `mapstructure:"scoring" json:"scoring" yaml:"scoring"`

	Jobs struct {
		SyncInterval string `mapstructure:"sync_interval" json:"sync_interval" yaml:"sync_interval"` // Default/fallback interval
		Mode         string `mapstructure:"mode" json:"mode" yaml:"mode"`                            // Default/fallback mode: "direct" or "jellyseerr"

		// Global limits
		GlobalLimitMovies int    `mapstructure:"global_limit_movies" json:"global_limit_movies" yaml:"global_limit_movies,omitempty"`
		GlobalLimitShows  int    `mapstructure:"global_limit_shows" json:"global_limit_shows" yaml:"global_limit_shows,omitempty"`
		GlobalPeriod      string `mapstructure:"global_period" json:"global_period" yaml:"global_period,omitempty"` // sync, daily, weekly, monthly

		TrendingMovies struct {
			Enabled             bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit               int    `mapstructure:"limit" json:"limit" yaml:"limit"`
			MinGlobalPicks      int    `mapstructure:"min_global_picks" json:"min_global_picks" yaml:"min_global_picks,omitempty"`
			SyncInterval        string `mapstructure:"sync_interval" json:"sync_interval" yaml:"sync_interval,omitempty"`
			Mode                string `mapstructure:"mode" json:"mode" yaml:"mode,omitempty"`
			MinimumAvailability string `mapstructure:"minimum_availability" json:"minimum_availability" yaml:"minimum_availability,omitempty"`
		} `mapstructure:"trending_movies" json:"trending_movies" yaml:"trending_movies"`

		TrendingShows struct {
			Enabled        bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit          int    `mapstructure:"limit" json:"limit" yaml:"limit"`
			MinGlobalPicks int    `mapstructure:"min_global_picks" json:"min_global_picks" yaml:"min_global_picks,omitempty"`
			SyncInterval   string `mapstructure:"sync_interval" json:"sync_interval" yaml:"sync_interval,omitempty"`
			Mode           string `mapstructure:"mode" json:"mode" yaml:"mode,omitempty"`
		} `mapstructure:"trending_shows" json:"trending_shows" yaml:"trending_shows"`

		PopularMovies struct {
			Enabled             bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit               int    `mapstructure:"limit" json:"limit" yaml:"limit"`
			MinGlobalPicks      int    `mapstructure:"min_global_picks" json:"min_global_picks" yaml:"min_global_picks,omitempty"`
			SyncInterval        string `mapstructure:"sync_interval" json:"sync_interval" yaml:"sync_interval,omitempty"`
			Mode                string `mapstructure:"mode" json:"mode" yaml:"mode,omitempty"`
			MinimumAvailability string `mapstructure:"minimum_availability" json:"minimum_availability" yaml:"minimum_availability,omitempty"`
		} `mapstructure:"popular_movies" json:"popular_movies" yaml:"popular_movies"`

		PopularShows struct {
			Enabled        bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit          int    `mapstructure:"limit" json:"limit" yaml:"limit"`
			MinGlobalPicks int    `mapstructure:"min_global_picks" json:"min_global_picks" yaml:"min_global_picks,omitempty"`
			SyncInterval   string `mapstructure:"sync_interval" json:"sync_interval" yaml:"sync_interval,omitempty"`
			Mode           string `mapstructure:"mode" json:"mode" yaml:"mode,omitempty"`
		} `mapstructure:"popular_shows" json:"popular_shows" yaml:"popular_shows"`

		BoxOffice struct {
			Enabled             bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit               int    `mapstructure:"limit" json:"limit" yaml:"limit"`
			MinGlobalPicks      int    `mapstructure:"min_global_picks" json:"min_global_picks" yaml:"min_global_picks,omitempty"`
			SyncInterval        string `mapstructure:"sync_interval" json:"sync_interval" yaml:"sync_interval,omitempty"`
			Mode                string `mapstructure:"mode" json:"mode" yaml:"mode,omitempty"`
			MinimumAvailability string `mapstructure:"minimum_availability" json:"minimum_availability" yaml:"minimum_availability,omitempty"`
		} `mapstructure:"box_office" json:"box_office" yaml:"box_office"`

		FavoritedMovies struct {
			Enabled             bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit               int    `mapstructure:"limit" json:"limit" yaml:"limit"`
			Period              string `mapstructure:"period" json:"period" yaml:"period"`
			MinGlobalPicks      int    `mapstructure:"min_global_picks" json:"min_global_picks" yaml:"min_global_picks,omitempty"`
			SyncInterval        string `mapstructure:"sync_interval" json:"sync_interval" yaml:"sync_interval,omitempty"`
			Mode                string `mapstructure:"mode" json:"mode" yaml:"mode,omitempty"`
			MinimumAvailability string `mapstructure:"minimum_availability" json:"minimum_availability" yaml:"minimum_availability,omitempty"`
		} `mapstructure:"favorited_movies" json:"favorited_movies" yaml:"favorited_movies"`

		PlayedMovies struct {
			Enabled             bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit               int    `mapstructure:"limit" json:"limit" yaml:"limit"`
			Period              string `mapstructure:"period" json:"period" yaml:"period"`
			MinGlobalPicks      int    `mapstructure:"min_global_picks" json:"min_global_picks" yaml:"min_global_picks,omitempty"`
			SyncInterval        string `mapstructure:"sync_interval" json:"sync_interval" yaml:"sync_interval,omitempty"`
			Mode                string `mapstructure:"mode" json:"mode" yaml:"mode,omitempty"`
			MinimumAvailability string `mapstructure:"minimum_availability" json:"minimum_availability" yaml:"minimum_availability,omitempty"`
		} `mapstructure:"played_movies" json:"played_movies" yaml:"played_movies"`

		WatchedMovies struct {
			Enabled             bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit               int    `mapstructure:"limit" json:"limit" yaml:"limit"`
			Period              string `mapstructure:"period" json:"period" yaml:"period"`
			MinGlobalPicks      int    `mapstructure:"min_global_picks" json:"min_global_picks" yaml:"min_global_picks,omitempty"`
			SyncInterval        string `mapstructure:"sync_interval" json:"sync_interval" yaml:"sync_interval,omitempty"`
			Mode                string `mapstructure:"mode" json:"mode" yaml:"mode,omitempty"`
			MinimumAvailability string `mapstructure:"minimum_availability" json:"minimum_availability" yaml:"minimum_availability,omitempty"`
		} `mapstructure:"watched_movies" json:"watched_movies" yaml:"watched_movies"`

		CollectedMovies struct {
			Enabled             bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit               int    `mapstructure:"limit" json:"limit" yaml:"limit"`
			Period              string `mapstructure:"period" json:"period" yaml:"period"`
			MinGlobalPicks      int    `mapstructure:"min_global_picks" json:"min_global_picks" yaml:"min_global_picks,omitempty"`
			SyncInterval        string `mapstructure:"sync_interval" json:"sync_interval" yaml:"sync_interval,omitempty"`
			Mode                string `mapstructure:"mode" json:"mode" yaml:"mode,omitempty"`
			MinimumAvailability string `mapstructure:"minimum_availability" json:"minimum_availability" yaml:"minimum_availability,omitempty"`
		} `mapstructure:"collected_movies" json:"collected_movies" yaml:"collected_movies"`

		AnticipatedMovies struct {
			Enabled             bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit               int    `mapstructure:"limit" json:"limit" yaml:"limit"`
			MinGlobalPicks      int    `mapstructure:"min_global_picks" json:"min_global_picks" yaml:"min_global_picks,omitempty"`
			SyncInterval        string `mapstructure:"sync_interval" json:"sync_interval" yaml:"sync_interval,omitempty"`
			Mode                string `mapstructure:"mode" json:"mode" yaml:"mode,omitempty"`
			MinimumAvailability string `mapstructure:"minimum_availability" json:"minimum_availability" yaml:"minimum_availability,omitempty"`
		} `mapstructure:"anticipated_movies" json:"anticipated_movies" yaml:"anticipated_movies"`

		FavoritedShows struct {
			Enabled        bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit          int    `mapstructure:"limit" json:"limit" yaml:"limit"`
			Period         string `mapstructure:"period" json:"period" yaml:"period"`
			MinGlobalPicks int    `mapstructure:"min_global_picks" json:"min_global_picks" yaml:"min_global_picks,omitempty"`
			SyncInterval   string `mapstructure:"sync_interval" json:"sync_interval" yaml:"sync_interval,omitempty"`
			Mode           string `mapstructure:"mode" json:"mode" yaml:"mode,omitempty"`
		} `mapstructure:"favorited_shows" json:"favorited_shows" yaml:"favorited_shows"`

		PlayedShows struct {
			Enabled        bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit          int    `mapstructure:"limit" json:"limit" yaml:"limit"`
			Period         string `mapstructure:"period" json:"period" yaml:"period"`
			MinGlobalPicks int    `mapstructure:"min_global_picks" json:"min_global_picks" yaml:"min_global_picks,omitempty"`
			SyncInterval   string `mapstructure:"sync_interval" json:"sync_interval" yaml:"sync_interval,omitempty"`
			Mode           string `mapstructure:"mode" json:"mode" yaml:"mode,omitempty"`
		} `mapstructure:"played_shows" json:"played_shows" yaml:"played_shows"`

		WatchedShows struct {
			Enabled        bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit          int    `mapstructure:"limit" json:"limit" yaml:"limit"`
			Period         string `mapstructure:"period" json:"period" yaml:"period"`
			MinGlobalPicks int    `mapstructure:"min_global_picks" json:"min_global_picks" yaml:"min_global_picks,omitempty"`
			SyncInterval   string `mapstructure:"sync_interval" json:"sync_interval" yaml:"sync_interval,omitempty"`
			Mode           string `mapstructure:"mode" json:"mode" yaml:"mode,omitempty"`
		} `mapstructure:"watched_shows" json:"watched_shows" yaml:"watched_shows"`

		CollectedShows struct {
			Enabled        bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit          int    `mapstructure:"limit" json:"limit" yaml:"limit"`
			Period         string `mapstructure:"period" json:"period" yaml:"period"`
			MinGlobalPicks int    `mapstructure:"min_global_picks" json:"min_global_picks" yaml:"min_global_picks,omitempty"`
			SyncInterval   string `mapstructure:"sync_interval" json:"sync_interval" yaml:"sync_interval,omitempty"`
			Mode           string `mapstructure:"mode" json:"mode" yaml:"mode,omitempty"`
		} `mapstructure:"collected_shows" json:"collected_shows" yaml:"collected_shows"`

		AnticipatedShows struct {
			Enabled        bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit          int    `mapstructure:"limit" json:"limit" yaml:"limit"`
			MinGlobalPicks int    `mapstructure:"min_global_picks" json:"min_global_picks" yaml:"min_global_picks,omitempty"`
			SyncInterval   string `mapstructure:"sync_interval" json:"sync_interval" yaml:"sync_interval,omitempty"`
			Mode           string `mapstructure:"mode" json:"mode" yaml:"mode,omitempty"`
		} `mapstructure:"anticipated_shows" json:"anticipated_shows" yaml:"anticipated_shows"`

		SmartPopularMovies struct {
			Enabled             bool    `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit               int     `mapstructure:"limit" json:"limit" yaml:"limit"`
			BaseMinRating       float64 `mapstructure:"base_min_rating" json:"base_min_rating" yaml:"base_min_rating"`
			AdjustmentFactor    float64 `mapstructure:"adjustment_factor" json:"adjustment_factor" yaml:"adjustment_factor"`
			SyncInterval        string  `mapstructure:"sync_interval" json:"sync_interval" yaml:"sync_interval,omitempty"`
			Mode                string  `mapstructure:"mode" json:"mode" yaml:"mode,omitempty"`
			MinimumAvailability string  `mapstructure:"minimum_availability" json:"minimum_availability" yaml:"minimum_availability,omitempty"`
		} `mapstructure:"smart_popular_movies" json:"smart_popular_movies" yaml:"smart_popular_movies"`

		SmartPopularShows struct {
			Enabled          bool    `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit            int     `mapstructure:"limit" json:"limit" yaml:"limit"`
			BaseMinRating    float64 `mapstructure:"base_min_rating" json:"base_min_rating" yaml:"base_min_rating"`
			AdjustmentFactor float64 `mapstructure:"adjustment_factor" json:"adjustment_factor" yaml:"adjustment_factor"`
			SyncInterval     string  `mapstructure:"sync_interval" json:"sync_interval" yaml:"sync_interval,omitempty"`
			Mode             string  `mapstructure:"mode" json:"mode" yaml:"mode,omitempty"`
		} `mapstructure:"smart_popular_shows" json:"smart_popular_shows" yaml:"smart_popular_shows"`
	} `mapstructure:"jobs" json:"jobs" yaml:"jobs"`

	Filters struct {
		Movies MovieFilters `mapstructure:"movies" json:"movies" yaml:"movies"`
		Shows  ShowFilters  `mapstructure:"shows" json:"shows" yaml:"shows"`
	} `mapstructure:"filters" json:"filters" yaml:"filters"`

	// Internal field to track config file path
	ConfigFilePath string `mapstructure:"-" json:"-" yaml:"-"`
}

// New creates a new Config instance with the given settings
func New(version string) (*Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Check if CONFIG_PATH environment variable is set
	if configPath := os.Getenv("CONFIG_PATH"); configPath != "" {
		v.AddConfigPath(configPath)
	}

	// Default config search paths
	v.AddConfigPath("./config")
	v.AddConfigPath("/home/nonroot/config")
	v.AddConfigPath("/app/config")
	v.AddConfigPath("/app/data")
	v.AddConfigPath(".")

	var configFileName string
	if version == "dev" {
		configFileName = "config.dev.yaml"
		v.SetConfigName(configFileName)
	} else {
		configFileName = "config.yaml"
		v.SetConfigName(configFileName)
	}

	var configFilePath string
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("No config file found, using only ENV variables")
			configFilePath = "" // No file found
		} else {
			return nil, fmt.Errorf("config load error: %w", err)
		}
	} else {
		configFilePath = v.ConfigFileUsed()
	}

	c := &Config{
		ConfigFilePath: configFilePath,
	}
	if err := v.Unmarshal(c); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Set default scoring values if not configured
	if c.Scoring.RatingScale == 0 {
		c.Scoring.RatingScale = 10
	}
	if c.Scoring.RecencyDays == 0 {
		c.Scoring.RecencyDays = 365
	}
	if c.Scoring.PopularityMetric == "" {
		c.Scoring.PopularityMetric = "votes"
	}
	// Set default weights if all are zero (indicates not configured)
	if c.Scoring.RatingWeight == 0 && c.Scoring.PopularityWeight == 0 && c.Scoring.RecencyWeight == 0 {
		c.Scoring.RatingWeight = 0.6
		c.Scoring.PopularityWeight = 0.3
		c.Scoring.RecencyWeight = 0.1
	}

	return c, nil
}

// Save writes the current config back to the config file
func (c *Config) Save() error {
	if c.ConfigFilePath == "" {
		return fmt.Errorf("no config file path set, cannot save")
	}

	// Marshal to YAML
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(c.ConfigFilePath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
