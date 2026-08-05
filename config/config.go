package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/mahcks/blockbusterr/pkg/enums"
	"gopkg.in/yaml.v3"
)

// MovieFilters represents filtering options for movies
type MovieFilters struct {
	AllowedCountries      []string `mapstructure:"allowed_countries" json:"allowed_countries" yaml:"allowed_countries"`
	AllowedLanguages      []string `mapstructure:"allowed_languages" json:"allowed_languages" yaml:"allowed_languages"`
	BlacklistedCountries  []string `mapstructure:"blacklisted_countries" json:"blacklisted_countries" yaml:"blacklisted_countries,omitempty"`
	BlacklistedLanguages  []string `mapstructure:"blacklisted_languages" json:"blacklisted_languages" yaml:"blacklisted_languages,omitempty"`
	BlacklistedGenres     []string `mapstructure:"blacklisted_genres" json:"blacklisted_genres" yaml:"blacklisted_genres"`
	BlacklistedKeywords   []string `mapstructure:"blacklisted_keywords" json:"blacklisted_keywords" yaml:"blacklisted_keywords"`
	RequiredGenres        []string `mapstructure:"required_genres" json:"required_genres" yaml:"required_genres,omitempty"`
	RequiredKeywords      []string `mapstructure:"required_keywords" json:"required_keywords" yaml:"required_keywords,omitempty"`
	AllowCountries        []string `mapstructure:"allow_countries" json:"allow_countries" yaml:"allow_countries,omitempty"`
	AllowLanguages        []string `mapstructure:"allow_languages" json:"allow_languages" yaml:"allow_languages,omitempty"`
	AllowGenres           []string `mapstructure:"allow_genres" json:"allow_genres" yaml:"allow_genres,omitempty"`
	AllowKeywords         []string `mapstructure:"allow_keywords" json:"allow_keywords" yaml:"allow_keywords,omitempty"`
	AllowMinRating        float64  `mapstructure:"allow_min_rating" json:"allow_min_rating" yaml:"allow_min_rating,omitempty"`
	BlacklistedTMDBIds    []int    `mapstructure:"blacklisted_tmdb_ids" json:"blacklisted_tmdb_ids" yaml:"blacklisted_tmdb_ids"`
	BlacklistedMinRuntime int      `mapstructure:"blacklisted_min_runtime" json:"blacklisted_min_runtime" yaml:"blacklisted_min_runtime"`
	BlacklistedMaxRuntime int      `mapstructure:"blacklisted_max_runtime" json:"blacklisted_max_runtime" yaml:"blacklisted_max_runtime"`
	BlacklistedMinYear    int      `mapstructure:"blacklisted_min_year" json:"blacklisted_min_year" yaml:"blacklisted_min_year"`
	BlacklistedMaxYear    int      `mapstructure:"blacklisted_max_year" json:"blacklisted_max_year" yaml:"blacklisted_max_year"`
	MinRating             float64  `mapstructure:"min_rating" json:"min_rating" yaml:"min_rating"`
	MinVotes              int      `mapstructure:"min_votes" json:"min_votes" yaml:"min_votes"`
	CertificationCountry  string   `mapstructure:"certification_country" json:"certification_country" yaml:"certification_country,omitempty"`
	AllowedCertifications []string `mapstructure:"allowed_certifications" json:"allowed_certifications" yaml:"allowed_certifications,omitempty"`
	BlockedCertifications []string `mapstructure:"blocked_certifications" json:"blocked_certifications" yaml:"blocked_certifications,omitempty"`
	UnknownCertification  string   `mapstructure:"unknown_certification" json:"unknown_certification" yaml:"unknown_certification"`
}

// ShowFilters represents filtering options for TV shows
type ShowFilters struct {
	AllowedCountries      []string `mapstructure:"allowed_countries" json:"allowed_countries" yaml:"allowed_countries"`
	AllowedLanguages      []string `mapstructure:"allowed_languages" json:"allowed_languages" yaml:"allowed_languages"`
	BlacklistedCountries  []string `mapstructure:"blacklisted_countries" json:"blacklisted_countries" yaml:"blacklisted_countries,omitempty"`
	BlacklistedLanguages  []string `mapstructure:"blacklisted_languages" json:"blacklisted_languages" yaml:"blacklisted_languages,omitempty"`
	BlacklistedGenres     []string `mapstructure:"blacklisted_genres" json:"blacklisted_genres" yaml:"blacklisted_genres"`
	BlacklistedKeywords   []string `mapstructure:"blacklisted_keywords" json:"blacklisted_keywords" yaml:"blacklisted_keywords"`
	BlacklistedNetworks   []string `mapstructure:"blacklisted_networks" json:"blacklisted_networks" yaml:"blacklisted_networks"`
	RequiredGenres        []string `mapstructure:"required_genres" json:"required_genres" yaml:"required_genres,omitempty"`
	RequiredKeywords      []string `mapstructure:"required_keywords" json:"required_keywords" yaml:"required_keywords,omitempty"`
	RequiredNetworks      []string `mapstructure:"required_networks" json:"required_networks" yaml:"required_networks,omitempty"`
	AllowCountries        []string `mapstructure:"allow_countries" json:"allow_countries" yaml:"allow_countries,omitempty"`
	AllowLanguages        []string `mapstructure:"allow_languages" json:"allow_languages" yaml:"allow_languages,omitempty"`
	AllowGenres           []string `mapstructure:"allow_genres" json:"allow_genres" yaml:"allow_genres,omitempty"`
	AllowKeywords         []string `mapstructure:"allow_keywords" json:"allow_keywords" yaml:"allow_keywords,omitempty"`
	AllowNetworks         []string `mapstructure:"allow_networks" json:"allow_networks" yaml:"allow_networks,omitempty"`
	AllowMinRating        float64  `mapstructure:"allow_min_rating" json:"allow_min_rating" yaml:"allow_min_rating,omitempty"`
	BlacklistedTVDBIds    []int    `mapstructure:"blacklisted_tvdb_ids" json:"blacklisted_tvdb_ids" yaml:"blacklisted_tvdb_ids"`
	BlacklistedMinRuntime int      `mapstructure:"blacklisted_min_runtime" json:"blacklisted_min_runtime" yaml:"blacklisted_min_runtime"`
	BlacklistedMaxRuntime int      `mapstructure:"blacklisted_max_runtime" json:"blacklisted_max_runtime" yaml:"blacklisted_max_runtime"`
	BlacklistedMinYear    int      `mapstructure:"blacklisted_min_year" json:"blacklisted_min_year" yaml:"blacklisted_min_year"`
	BlacklistedMaxYear    int      `mapstructure:"blacklisted_max_year" json:"blacklisted_max_year" yaml:"blacklisted_max_year"`
	MinRating             float64  `mapstructure:"min_rating" json:"min_rating" yaml:"min_rating"`
	MinVotes              int      `mapstructure:"min_votes" json:"min_votes" yaml:"min_votes"`
	CertificationCountry  string   `mapstructure:"certification_country" json:"certification_country" yaml:"certification_country,omitempty"`
	AllowedCertifications []string `mapstructure:"allowed_certifications" json:"allowed_certifications" yaml:"allowed_certifications,omitempty"`
	BlockedCertifications []string `mapstructure:"blocked_certifications" json:"blocked_certifications" yaml:"blocked_certifications,omitempty"`
	UnknownCertification  string   `mapstructure:"unknown_certification" json:"unknown_certification" yaml:"unknown_certification"`
}

// FilterConfig contains the movie and show policies used during discovery.
type FilterConfig struct {
	Movies MovieFilters `mapstructure:"movies" json:"movies" yaml:"movies"`
	Shows  ShowFilters  `mapstructure:"shows" json:"shows" yaml:"shows"`
}

const (
	DefaultMoviesRuleSetID = "default-movies"
	DefaultShowsRuleSetID  = "default-shows"
)

// RuleSet is a reusable, media-specific discovery policy.
type RuleSet struct {
	ID       string        `mapstructure:"id" json:"id" yaml:"id"`
	Name     string        `mapstructure:"name" json:"name" yaml:"name"`
	Media    string        `mapstructure:"media" json:"media" yaml:"media"`
	Revision int           `mapstructure:"revision" json:"revision" yaml:"revision"`
	Movies   *MovieFilters `mapstructure:"movies" json:"movies,omitempty" yaml:"movies,omitempty"`
	Shows    *ShowFilters  `mapstructure:"shows" json:"shows,omitempty" yaml:"shows,omitempty"`
}

type TitleExceptions struct {
	AllowedMovieTMDBIDs []int `mapstructure:"allowed_movie_tmdb_ids" json:"allowed_movie_tmdb_ids" yaml:"allowed_movie_tmdb_ids,omitempty"`
	BlockedMovieTMDBIDs []int `mapstructure:"blocked_movie_tmdb_ids" json:"blocked_movie_tmdb_ids" yaml:"blocked_movie_tmdb_ids,omitempty"`
	AllowedShowTVDBIDs  []int `mapstructure:"allowed_show_tvdb_ids" json:"allowed_show_tvdb_ids" yaml:"allowed_show_tvdb_ids,omitempty"`
	BlockedShowTVDBIDs  []int `mapstructure:"blocked_show_tvdb_ids" json:"blocked_show_tvdb_ids" yaml:"blocked_show_tvdb_ids,omitempty"`
}

// ListLocator identifies a provider-owned list without storing an arbitrary URL.
type ListLocator struct {
	Kind     string `mapstructure:"kind" json:"kind" yaml:"kind"`
	Owner    string `mapstructure:"owner" json:"owner,omitempty" yaml:"owner,omitempty"`
	ListID   string `mapstructure:"list_id" json:"list_id,omitempty" yaml:"list_id,omitempty"`
	Slug     string `mapstructure:"slug" json:"slug,omitempty" yaml:"slug,omitempty"`
	Ordering string `mapstructure:"ordering" json:"ordering,omitempty" yaml:"ordering,omitempty"`
}

type RecommendationSeedList struct {
	Source string      `mapstructure:"source" json:"source" yaml:"source"`
	List   ListLocator `mapstructure:"list" json:"list" yaml:"list"`
}

// DynamicJob represents a user-defined job instance that can be created, modified, and deleted
type DynamicJob struct {
	ID                  string                  `mapstructure:"id" json:"id" yaml:"id"`
	Name                string                  `mapstructure:"name" json:"name" yaml:"name"`
	Enabled             bool                    `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	Type                string                  `mapstructure:"type" json:"type" yaml:"type"`                                                           // Job type: trending, popular, watched, collected, favorited, played, anticipated, box_office, smart_popular
	Source              string                  `mapstructure:"source" json:"source" yaml:"source"`                                                     // Discovery source: trakt, tmdb, or simkl
	MediaType           string                  `mapstructure:"media" json:"media" yaml:"media"`                                                        // Media type: movie or show
	Limit               int                     `mapstructure:"limit" json:"limit" yaml:"limit"`                                                        // Number of items to fetch
	Period              string                  `mapstructure:"period" json:"period" yaml:"period,omitempty"`                                           // Time period for watched/collected/favorited/played: weekly, monthly, yearly, all
	SyncInterval        string                  `mapstructure:"sync_interval" json:"sync_interval" yaml:"sync_interval,omitempty"`                      // Custom sync interval (overrides global)
	Mode                string                  `mapstructure:"mode" json:"mode" yaml:"mode,omitempty"`                                                 // Execution mode: direct or jellyseerr (overrides global)
	MinimumAvailability string                  `mapstructure:"minimum_availability" json:"minimum_availability" yaml:"minimum_availability,omitempty"` // For Radarr: announced, in_cinemas, released
	Monitor             string                  `mapstructure:"monitor" json:"monitor" yaml:"monitor,omitempty"`                                        // Monitor setting for Radarr/Sonarr
	BaseMinRating       float64                 `mapstructure:"base_min_rating" json:"base_min_rating" yaml:"base_min_rating,omitempty"`                // For smart jobs: base minimum rating
	AdjustmentFactor    float64                 `mapstructure:"adjustment_factor" json:"adjustment_factor" yaml:"adjustment_factor,omitempty"`          // For smart jobs: rating adjustment factor
	DeliveryLimit       int                     `mapstructure:"delivery_limit" json:"delivery_limit" yaml:"delivery_limit,omitempty"`                   // Maximum successful deliveries per run; zero is unlimited
	SelectionCycle      bool                    `mapstructure:"selection_cycle" json:"selection_cycle" yaml:"selection_cycle,omitempty"`
	MinimumPicks        int                     `mapstructure:"minimum_picks" json:"minimum_picks" yaml:"minimum_picks,omitempty"`
	RepeatPolicy        string                  `mapstructure:"repeat_policy" json:"repeat_policy" yaml:"repeat_policy,omitempty"`
	UseCustomFilters    bool                    `mapstructure:"use_custom_filters" json:"use_custom_filters" yaml:"use_custom_filters,omitempty"`
	Filters             FilterConfig            `mapstructure:"filters" json:"filters" yaml:"filters,omitempty"`
	RuleSetID           string                  `mapstructure:"rule_set_id" json:"rule_set_id" yaml:"rule_set_id,omitempty"`
	List                *ListLocator            `mapstructure:"list" json:"list,omitempty" yaml:"list,omitempty"`
	RecommendationSeeds []int                   `mapstructure:"recommendation_seeds" json:"recommendation_seeds,omitempty" yaml:"recommendation_seeds,omitempty"`
	RecommendationList  *RecommendationSeedList `mapstructure:"recommendation_list" json:"recommendation_list,omitempty" yaml:"recommendation_list,omitempty"`
	SeriesType          string                  `mapstructure:"series_type" json:"series_type,omitempty" yaml:"series_type,omitempty"`
}

// Config represents the application configuration
type Config struct {
	Version string `mapstructure:"version" yaml:"version,omitempty"`

	Trakt struct {
		ClientID     string `mapstructure:"client_id" json:"client_id" yaml:"client_id"`
		ClientSecret string `mapstructure:"client_secret" json:"-" yaml:"client_secret"`
		AccessToken  string `mapstructure:"access_token" json:"-" yaml:"access_token,omitempty"`
		RefreshToken string `mapstructure:"refresh_token" json:"-" yaml:"refresh_token,omitempty"`
		TokenExpires int64  `mapstructure:"token_expires" json:"-" yaml:"token_expires,omitempty"`
	} `mapstructure:"trakt" json:"trakt" yaml:"trakt"`

	TMDB struct {
		APIKey    string `mapstructure:"api_key" json:"-" yaml:"api_key"`
		SessionID string `mapstructure:"session_id" json:"-" yaml:"session_id,omitempty"`
		AccountID int    `mapstructure:"account_id" json:"account_id,omitempty" yaml:"account_id,omitempty"`
	} `mapstructure:"tmdb" json:"tmdb" yaml:"tmdb"`

	Simkl struct {
		ClientID string `mapstructure:"client_id" json:"client_id" yaml:"client_id"`
	} `mapstructure:"simkl" json:"simkl" yaml:"simkl"`

	MDBList struct {
		APIKey string `mapstructure:"api_key" json:"-" yaml:"api_key"`
	} `mapstructure:"mdblist" json:"mdblist" yaml:"mdblist"`

	Letterboxd struct {
		ExperimentalScraping bool `mapstructure:"experimental_scraping" json:"experimental_scraping" yaml:"experimental_scraping"`
	} `mapstructure:"letterboxd" json:"letterboxd" yaml:"letterboxd"`

	Radarr struct {
		URL                 string `mapstructure:"url" json:"url" yaml:"url"`
		APIKey              string `mapstructure:"api_key" json:"-" yaml:"api_key"`
		QualityProfile      int    `mapstructure:"quality_profile" json:"quality_profile" yaml:"quality_profile"`
		RootFolder          string `mapstructure:"root_folder" json:"root_folder" yaml:"root_folder"`
		MinimumAvailability string `mapstructure:"minimum_availability" json:"minimum_availability" yaml:"minimum_availability,omitempty"`
		Monitor             string `mapstructure:"monitor" json:"monitor" yaml:"monitor,omitempty"`
	} `mapstructure:"radarr" json:"radarr" yaml:"radarr"`

	Sonarr struct {
		URL            string `mapstructure:"url" json:"url" yaml:"url"`
		APIKey         string `mapstructure:"api_key" json:"-" yaml:"api_key"`
		QualityProfile int    `mapstructure:"quality_profile" json:"quality_profile" yaml:"quality_profile"`
		RootFolder     string `mapstructure:"root_folder" json:"root_folder" yaml:"root_folder"`
		Monitor        string `mapstructure:"monitor" json:"monitor" yaml:"monitor"`
	} `mapstructure:"sonarr" json:"sonarr" yaml:"sonarr"`

	Jellyseerr struct {
		URL                string `mapstructure:"url" json:"url" yaml:"url"`
		APIKey             string `mapstructure:"api_key" json:"-" yaml:"api_key"`
		UserID             string `mapstructure:"user_id" json:"user_id" yaml:"user_id"` // Optional: request as specific user
		RequestCredentials struct {
			Email    string `mapstructure:"email" json:"email" yaml:"email"`
			Password string `mapstructure:"password" json:"-" yaml:"password"`
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
		SyncInterval      string `mapstructure:"sync_interval" json:"sync_interval" yaml:"sync_interval"` // Default/fallback interval
		Mode              string `mapstructure:"mode" json:"mode" yaml:"mode"`                            // Default/fallback mode: "direct" or "jellyseerr"
		GlobalLimitMovies int    `mapstructure:"global_limit_movies" json:"global_limit_movies" yaml:"global_limit_movies,omitempty"`
		GlobalLimitShows  int    `mapstructure:"global_limit_shows" json:"global_limit_shows" yaml:"global_limit_shows,omitempty"`
		GlobalPeriod      string `mapstructure:"global_period" json:"global_period" yaml:"global_period,omitempty"` // daily, weekly, monthly
		RepeatPolicy      string `mapstructure:"repeat_policy" json:"repeat_policy" yaml:"repeat_policy,omitempty"`
		Selection         struct {
			Enabled      bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			SyncInterval string `mapstructure:"sync_interval" json:"sync_interval" yaml:"sync_interval,omitempty"`
			MovieLimit   int    `mapstructure:"movie_limit" json:"movie_limit" yaml:"movie_limit,omitempty"`
			ShowLimit    int    `mapstructure:"show_limit" json:"show_limit" yaml:"show_limit,omitempty"`
		} `mapstructure:"selection" json:"selection" yaml:"selection"`

		// Dynamic job list (new format - allows multiple instances of same job type)
		List []DynamicJob `mapstructure:"list" json:"list" yaml:"list,omitempty"`

		TrendingMovies struct {
			Enabled             bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit               int    `mapstructure:"limit" json:"limit" yaml:"limit"`
			SyncInterval        string `mapstructure:"sync_interval" json:"sync_interval" yaml:"sync_interval,omitempty"`
			Mode                string `mapstructure:"mode" json:"mode" yaml:"mode,omitempty"`
			MinimumAvailability string `mapstructure:"minimum_availability" json:"minimum_availability" yaml:"minimum_availability,omitempty"`
			Monitor             string `mapstructure:"monitor" json:"monitor" yaml:"monitor,omitempty"` // "movieOnly", "movieAndCollection", or "none"
		} `mapstructure:"trending_movies" json:"trending_movies" yaml:"trending_movies"`

		TrendingShows struct {
			Enabled      bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit        int    `mapstructure:"limit" json:"limit" yaml:"limit"`
			SyncInterval string `mapstructure:"sync_interval" json:"sync_interval" yaml:"sync_interval,omitempty"`
			Mode         string `mapstructure:"mode" json:"mode" yaml:"mode,omitempty"`
			Monitor      string `mapstructure:"monitor" json:"monitor" yaml:"monitor,omitempty"`
		} `mapstructure:"trending_shows" json:"trending_shows" yaml:"trending_shows"`

		PopularMovies struct {
			Enabled             bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit               int    `mapstructure:"limit" json:"limit" yaml:"limit"`
			SyncInterval        string `mapstructure:"sync_interval" json:"sync_interval" yaml:"sync_interval,omitempty"`
			Mode                string `mapstructure:"mode" json:"mode" yaml:"mode,omitempty"`
			MinimumAvailability string `mapstructure:"minimum_availability" json:"minimum_availability" yaml:"minimum_availability,omitempty"`
			Monitor             string `mapstructure:"monitor" json:"monitor" yaml:"monitor,omitempty"` // "movieOnly", "movieAndCollection", or "none"
		} `mapstructure:"popular_movies" json:"popular_movies" yaml:"popular_movies"`

		PopularShows struct {
			Enabled      bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit        int    `mapstructure:"limit" json:"limit" yaml:"limit"`
			SyncInterval string `mapstructure:"sync_interval" json:"sync_interval" yaml:"sync_interval,omitempty"`
			Mode         string `mapstructure:"mode" json:"mode" yaml:"mode,omitempty"`
			Monitor      string `mapstructure:"monitor" json:"monitor" yaml:"monitor,omitempty"`
		} `mapstructure:"popular_shows" json:"popular_shows" yaml:"popular_shows"`

		BoxOffice struct {
			Enabled             bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit               int    `mapstructure:"limit" json:"limit" yaml:"limit"`
			SyncInterval        string `mapstructure:"sync_interval" json:"sync_interval" yaml:"sync_interval,omitempty"`
			Mode                string `mapstructure:"mode" json:"mode" yaml:"mode,omitempty"`
			MinimumAvailability string `mapstructure:"minimum_availability" json:"minimum_availability" yaml:"minimum_availability,omitempty"`
			Monitor             string `mapstructure:"monitor" json:"monitor" yaml:"monitor,omitempty"` // "movieOnly", "movieAndCollection", or "none"
		} `mapstructure:"box_office" json:"box_office" yaml:"box_office"`

		FavoritedMovies struct {
			Enabled             bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit               int    `mapstructure:"limit" json:"limit" yaml:"limit"`
			Period              string `mapstructure:"period" json:"period" yaml:"period"`
			SyncInterval        string `mapstructure:"sync_interval" json:"sync_interval" yaml:"sync_interval,omitempty"`
			Mode                string `mapstructure:"mode" json:"mode" yaml:"mode,omitempty"`
			MinimumAvailability string `mapstructure:"minimum_availability" json:"minimum_availability" yaml:"minimum_availability,omitempty"`
			Monitor             string `mapstructure:"monitor" json:"monitor" yaml:"monitor,omitempty"` // "movieOnly", "movieAndCollection", or "none"
		} `mapstructure:"favorited_movies" json:"favorited_movies" yaml:"favorited_movies"`

		PlayedMovies struct {
			Enabled             bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit               int    `mapstructure:"limit" json:"limit" yaml:"limit"`
			Period              string `mapstructure:"period" json:"period" yaml:"period"`
			SyncInterval        string `mapstructure:"sync_interval" json:"sync_interval" yaml:"sync_interval,omitempty"`
			Mode                string `mapstructure:"mode" json:"mode" yaml:"mode,omitempty"`
			MinimumAvailability string `mapstructure:"minimum_availability" json:"minimum_availability" yaml:"minimum_availability,omitempty"`
			Monitor             string `mapstructure:"monitor" json:"monitor" yaml:"monitor,omitempty"` // "movieOnly", "movieAndCollection", or "none"
		} `mapstructure:"played_movies" json:"played_movies" yaml:"played_movies"`

		WatchedMovies struct {
			Enabled             bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit               int    `mapstructure:"limit" json:"limit" yaml:"limit"`
			Period              string `mapstructure:"period" json:"period" yaml:"period"`
			SyncInterval        string `mapstructure:"sync_interval" json:"sync_interval" yaml:"sync_interval,omitempty"`
			Mode                string `mapstructure:"mode" json:"mode" yaml:"mode,omitempty"`
			MinimumAvailability string `mapstructure:"minimum_availability" json:"minimum_availability" yaml:"minimum_availability,omitempty"`
			Monitor             string `mapstructure:"monitor" json:"monitor" yaml:"monitor,omitempty"` // "movieOnly", "movieAndCollection", or "none"
		} `mapstructure:"watched_movies" json:"watched_movies" yaml:"watched_movies"`

		CollectedMovies struct {
			Enabled             bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit               int    `mapstructure:"limit" json:"limit" yaml:"limit"`
			Period              string `mapstructure:"period" json:"period" yaml:"period"`
			SyncInterval        string `mapstructure:"sync_interval" json:"sync_interval" yaml:"sync_interval,omitempty"`
			Mode                string `mapstructure:"mode" json:"mode" yaml:"mode,omitempty"`
			MinimumAvailability string `mapstructure:"minimum_availability" json:"minimum_availability" yaml:"minimum_availability,omitempty"`
			Monitor             string `mapstructure:"monitor" json:"monitor" yaml:"monitor,omitempty"` // "movieOnly", "movieAndCollection", or "none"
		} `mapstructure:"collected_movies" json:"collected_movies" yaml:"collected_movies"`

		AnticipatedMovies struct {
			Enabled             bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit               int    `mapstructure:"limit" json:"limit" yaml:"limit"`
			SyncInterval        string `mapstructure:"sync_interval" json:"sync_interval" yaml:"sync_interval,omitempty"`
			Mode                string `mapstructure:"mode" json:"mode" yaml:"mode,omitempty"`
			MinimumAvailability string `mapstructure:"minimum_availability" json:"minimum_availability" yaml:"minimum_availability,omitempty"`
			Monitor             string `mapstructure:"monitor" json:"monitor" yaml:"monitor,omitempty"` // "movieOnly", "movieAndCollection", or "none"
		} `mapstructure:"anticipated_movies" json:"anticipated_movies" yaml:"anticipated_movies"`

		FavoritedShows struct {
			Enabled      bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit        int    `mapstructure:"limit" json:"limit" yaml:"limit"`
			Period       string `mapstructure:"period" json:"period" yaml:"period"`
			SyncInterval string `mapstructure:"sync_interval" json:"sync_interval" yaml:"sync_interval,omitempty"`
			Mode         string `mapstructure:"mode" json:"mode" yaml:"mode,omitempty"`
			Monitor      string `mapstructure:"monitor" json:"monitor" yaml:"monitor,omitempty"`
		} `mapstructure:"favorited_shows" json:"favorited_shows" yaml:"favorited_shows"`

		PlayedShows struct {
			Enabled      bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit        int    `mapstructure:"limit" json:"limit" yaml:"limit"`
			Period       string `mapstructure:"period" json:"period" yaml:"period"`
			SyncInterval string `mapstructure:"sync_interval" json:"sync_interval" yaml:"sync_interval,omitempty"`
			Mode         string `mapstructure:"mode" json:"mode" yaml:"mode,omitempty"`
			Monitor      string `mapstructure:"monitor" json:"monitor" yaml:"monitor,omitempty"`
		} `mapstructure:"played_shows" json:"played_shows" yaml:"played_shows"`

		WatchedShows struct {
			Enabled      bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit        int    `mapstructure:"limit" json:"limit" yaml:"limit"`
			Period       string `mapstructure:"period" json:"period" yaml:"period"`
			SyncInterval string `mapstructure:"sync_interval" json:"sync_interval" yaml:"sync_interval,omitempty"`
			Mode         string `mapstructure:"mode" json:"mode" yaml:"mode,omitempty"`
			Monitor      string `mapstructure:"monitor" json:"monitor" yaml:"monitor,omitempty"`
		} `mapstructure:"watched_shows" json:"watched_shows" yaml:"watched_shows"`

		CollectedShows struct {
			Enabled      bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit        int    `mapstructure:"limit" json:"limit" yaml:"limit"`
			Period       string `mapstructure:"period" json:"period" yaml:"period"`
			SyncInterval string `mapstructure:"sync_interval" json:"sync_interval" yaml:"sync_interval,omitempty"`
			Mode         string `mapstructure:"mode" json:"mode" yaml:"mode,omitempty"`
			Monitor      string `mapstructure:"monitor" json:"monitor" yaml:"monitor,omitempty"`
		} `mapstructure:"collected_shows" json:"collected_shows" yaml:"collected_shows"`

		AnticipatedShows struct {
			Enabled      bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit        int    `mapstructure:"limit" json:"limit" yaml:"limit"`
			SyncInterval string `mapstructure:"sync_interval" json:"sync_interval" yaml:"sync_interval,omitempty"`
			Mode         string `mapstructure:"mode" json:"mode" yaml:"mode,omitempty"`
			Monitor      string `mapstructure:"monitor" json:"monitor" yaml:"monitor,omitempty"`
		} `mapstructure:"anticipated_shows" json:"anticipated_shows" yaml:"anticipated_shows"`

		SmartPopularMovies struct {
			Enabled             bool    `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit               int     `mapstructure:"limit" json:"limit" yaml:"limit"`
			BaseMinRating       float64 `mapstructure:"base_min_rating" json:"base_min_rating" yaml:"base_min_rating"`
			AdjustmentFactor    float64 `mapstructure:"adjustment_factor" json:"adjustment_factor" yaml:"adjustment_factor"`
			SyncInterval        string  `mapstructure:"sync_interval" json:"sync_interval" yaml:"sync_interval,omitempty"`
			Mode                string  `mapstructure:"mode" json:"mode" yaml:"mode,omitempty"`
			MinimumAvailability string  `mapstructure:"minimum_availability" json:"minimum_availability" yaml:"minimum_availability,omitempty"`
			Monitor             string  `mapstructure:"monitor" json:"monitor" yaml:"monitor,omitempty"` // "movieOnly", "movieAndCollection", or "none"
		} `mapstructure:"smart_popular_movies" json:"smart_popular_movies" yaml:"smart_popular_movies"`

		SmartPopularShows struct {
			Enabled          bool    `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
			Limit            int     `mapstructure:"limit" json:"limit" yaml:"limit"`
			BaseMinRating    float64 `mapstructure:"base_min_rating" json:"base_min_rating" yaml:"base_min_rating"`
			AdjustmentFactor float64 `mapstructure:"adjustment_factor" json:"adjustment_factor" yaml:"adjustment_factor"`
			SyncInterval     string  `mapstructure:"sync_interval" json:"sync_interval" yaml:"sync_interval,omitempty"`
			Mode             string  `mapstructure:"mode" json:"mode" yaml:"mode,omitempty"`
			Monitor          string  `mapstructure:"monitor" json:"monitor" yaml:"monitor,omitempty"`
		} `mapstructure:"smart_popular_shows" json:"smart_popular_shows" yaml:"smart_popular_shows"`
	} `mapstructure:"jobs" json:"jobs" yaml:"jobs"`

	Filters         FilterConfig    `mapstructure:"filters" json:"filters" yaml:"filters"`
	RuleSets        []RuleSet       `mapstructure:"rule_sets" json:"rule_sets" yaml:"rule_sets,omitempty"`
	TitleExceptions TitleExceptions `mapstructure:"title_exceptions" json:"title_exceptions" yaml:"title_exceptions,omitempty"`

	// Internal field to track config file path
	ConfigFilePath   string                                            `mapstructure:"-" json:"-" yaml:"-"`
	UpdateTraktToken func(access, refresh string, expires int64) error `mapstructure:"-" json:"-" yaml:"-"`
}

// Clone returns an independent configuration value suitable for copy-on-write updates.
func (c *Config) Clone() (*Config, error) {
	data, err := yaml.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("failed to clone config: %w", err)
	}
	clone := &Config{}
	if err := yaml.Unmarshal(data, clone); err != nil {
		return nil, fmt.Errorf("failed to clone config: %w", err)
	}
	clone.ConfigFilePath = c.ConfigFilePath
	clone.UpdateTraktToken = c.UpdateTraktToken
	return clone, nil
}

// New creates a new Config instance with the given settings
func New(version string) (*Config, error) {
	configFileName := "config.yaml"
	if version == "dev" {
		configFileName = "config.dev.yaml"
	}

	paths := []string{"./config", "../config", "../../config", "/home/nonroot/config", "/app/config", "/app/data", ".", "..", "../data"}
	if configPath := strings.TrimSpace(os.Getenv("CONFIG_PATH")); configPath != "" {
		paths = append([]string{configPath}, paths...)
	}

	c := &Config{}
	for _, dir := range paths {
		path := filepath.Join(dir, configFileName)
		data, err := os.ReadFile(path)
		if err == nil {
			if err := yaml.Unmarshal(data, c); err != nil {
				return nil, fmt.Errorf("config load error: %w", err)
			}
			c.ConfigFilePath = path
			break
		}
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("config load error: %w", err)
		}
	}
	if c.ConfigFilePath == "" {
		fmt.Println("No config file found, using only ENV variables")
	}
	if err := applyEnvironment(reflect.ValueOf(c).Elem(), nil); err != nil {
		return nil, fmt.Errorf("config environment: %w", err)
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
	if c.Jobs.GlobalPeriod != "daily" && c.Jobs.GlobalPeriod != "weekly" && c.Jobs.GlobalPeriod != "monthly" {
		c.Jobs.GlobalPeriod = "daily"
	}
	if !enums.RepeatPolicy(c.Jobs.RepeatPolicy).IsValid(false) {
		c.Jobs.RepeatPolicy = string(enums.RepeatPolicy90Days)
	}
	if c.Jobs.Selection.SyncInterval == "" {
		c.Jobs.Selection.SyncInterval = "24h"
	}
	c.MigrateRuleSets()

	return c, nil
}

func applyEnvironment(value reflect.Value, path []string) error {
	typeOfValue := value.Type()
	for i := 0; i < value.NumField(); i++ {
		field, fieldType := value.Field(i), typeOfValue.Field(i)
		name := strings.Split(fieldType.Tag.Get("yaml"), ",")[0]
		if name == "" || name == "-" || !field.CanSet() {
			continue
		}
		fieldPath := append(path, name)
		if field.Kind() == reflect.Struct {
			if err := applyEnvironment(field, fieldPath); err != nil {
				return err
			}
			continue
		}
		raw, ok := os.LookupEnv(strings.ToUpper(strings.Join(fieldPath, "_")))
		if !ok {
			continue
		}
		var err error
		switch field.Kind() {
		case reflect.String:
			field.SetString(raw)
		case reflect.Bool:
			var parsed bool
			parsed, err = strconv.ParseBool(raw)
			field.SetBool(parsed)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			var parsed int64
			parsed, err = strconv.ParseInt(raw, 10, field.Type().Bits())
			field.SetInt(parsed)
		case reflect.Float32, reflect.Float64:
			var parsed float64
			parsed, err = strconv.ParseFloat(raw, field.Type().Bits())
			field.SetFloat(parsed)
		default:
			err = yaml.Unmarshal([]byte(raw), field.Addr().Interface())
		}
		if err != nil {
			return fmt.Errorf("%s: %w", strings.ToUpper(strings.Join(fieldPath, "_")), err)
		}
	}
	return nil
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

	dir := filepath.Dir(c.ConfigFilePath)
	tmp, err := os.CreateTemp(dir, ".blockbusterr-config-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary config: %w", err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()
	mode := os.FileMode(0o644)
	if info, statErr := os.Stat(c.ConfigFilePath); statErr == nil {
		mode = info.Mode().Perm()
	}
	if err = tmp.Chmod(mode); err == nil {
		_, err = tmp.Write(data)
	}
	if err == nil {
		err = tmp.Sync()
	}
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return fmt.Errorf("failed to write temporary config: %w", err)
	}
	if err := os.Rename(tmpName, c.ConfigFilePath); err != nil {
		return fmt.Errorf("failed to replace config file: %w", err)
	}
	return nil
}

func defaultRuleSetID(media string) string {
	if media == "show" {
		return DefaultShowsRuleSetID
	}
	return DefaultMoviesRuleSetID
}

func (c *Config) RuleSetByID(id string) (*RuleSet, bool) {
	for i := range c.RuleSets {
		if c.RuleSets[i].ID == id {
			return &c.RuleSets[i], true
		}
	}
	return nil, false
}

func (c *Config) ResolveRuleSet(job DynamicJob) (*RuleSet, FilterConfig, error) {
	if len(c.RuleSets) == 0 {
		c.MigrateRuleSets()
	}
	id := job.RuleSetID
	if id == "" {
		id = defaultRuleSetID(job.MediaType)
	}
	rules, ok := c.RuleSetByID(id)
	if !ok {
		return nil, FilterConfig{}, fmt.Errorf("rule set %q was not found", id)
	}
	if rules.Media != job.MediaType {
		return nil, FilterConfig{}, fmt.Errorf("rule set %q is for %s jobs, not %s", rules.Name, rules.Media, job.MediaType)
	}
	effective := FilterConfig{}
	if job.MediaType == "movie" && rules.Movies != nil {
		effective.Movies = *rules.Movies
	} else if job.MediaType == "show" && rules.Shows != nil {
		effective.Shows = *rules.Shows
	} else {
		return nil, FilterConfig{}, fmt.Errorf("rule set %q has no %s rules", rules.Name, job.MediaType)
	}
	effective.Movies.BlacklistedTMDBIds = append(effective.Movies.BlacklistedTMDBIds, c.TitleExceptions.BlockedMovieTMDBIDs...)
	effective.Shows.BlacklistedTVDBIds = append(effective.Shows.BlacklistedTVDBIds, c.TitleExceptions.BlockedShowTVDBIDs...)
	return rules, effective, nil
}

func (c *Config) ValidateRuleSet(candidate RuleSet, exceptID string) error {
	candidate.Name = strings.TrimSpace(candidate.Name)
	if candidate.ID == "" || candidate.Name == "" {
		return fmt.Errorf("rule set ID and name are required")
	}
	if candidate.Media != "movie" && candidate.Media != "show" {
		return fmt.Errorf("media must be movie or show")
	}
	if candidate.Media == "movie" && (candidate.Movies == nil || candidate.Shows != nil) {
		return fmt.Errorf("movie rule sets require only a movie filter payload")
	}
	if candidate.Media == "show" && (candidate.Shows == nil || candidate.Movies != nil) {
		return fmt.Errorf("show rule sets require only a show filter payload")
	}
	var minYear, maxYear, minRuntime, maxRuntime, minVotes int
	var minRating, allowMinRating float64
	var certificationCountry, unknownCertification string
	var allowedCertifications, blockedCertifications []string
	if candidate.Media == "movie" {
		minYear, maxYear, minRuntime, maxRuntime, minRating, minVotes = candidate.Movies.BlacklistedMinYear, candidate.Movies.BlacklistedMaxYear, candidate.Movies.BlacklistedMinRuntime, candidate.Movies.BlacklistedMaxRuntime, candidate.Movies.MinRating, candidate.Movies.MinVotes
		allowMinRating = candidate.Movies.AllowMinRating
		certificationCountry, unknownCertification = candidate.Movies.CertificationCountry, candidate.Movies.UnknownCertification
		allowedCertifications, blockedCertifications = candidate.Movies.AllowedCertifications, candidate.Movies.BlockedCertifications
	} else {
		minYear, maxYear, minRuntime, maxRuntime, minRating, minVotes = candidate.Shows.BlacklistedMinYear, candidate.Shows.BlacklistedMaxYear, candidate.Shows.BlacklistedMinRuntime, candidate.Shows.BlacklistedMaxRuntime, candidate.Shows.MinRating, candidate.Shows.MinVotes
		allowMinRating = candidate.Shows.AllowMinRating
		certificationCountry, unknownCertification = candidate.Shows.CertificationCountry, candidate.Shows.UnknownCertification
		allowedCertifications, blockedCertifications = candidate.Shows.AllowedCertifications, candidate.Shows.BlockedCertifications
	}
	if minYear < 0 || maxYear < 0 || minRuntime < 0 || maxRuntime < 0 || minVotes < 0 {
		return fmt.Errorf("numeric rule values cannot be negative")
	}
	if minRating < 0 || minRating > 10 {
		return fmt.Errorf("minimum rating must be between 0 and 10")
	}
	if allowMinRating < 0 || allowMinRating > 10 {
		return fmt.Errorf("allow override rating must be between 0 and 10")
	}
	if minYear > 0 && maxYear > 0 && minYear > maxYear {
		return fmt.Errorf("minimum year cannot exceed maximum year")
	}
	if minRuntime > 0 && maxRuntime > 0 && minRuntime > maxRuntime {
		return fmt.Errorf("minimum runtime cannot exceed maximum runtime")
	}
	if unknownCertification == "" {
		unknownCertification = string(enums.CertificationUnknownAllow)
	}
	if !enums.CertificationUnknownPolicy(unknownCertification).Valid() {
		return fmt.Errorf("unknown certification policy must be allow or reject")
	}
	if len(allowedCertifications)+len(blockedCertifications) > 0 && len(strings.TrimSpace(certificationCountry)) != 2 {
		return fmt.Errorf("certification country must be a two-letter country code")
	}
	for _, rules := range c.RuleSets {
		if rules.ID != exceptID && strings.EqualFold(rules.Name, candidate.Name) && rules.Media == candidate.Media {
			return fmt.Errorf("a %s rule set named %q already exists", candidate.Media, candidate.Name)
		}
	}
	return nil
}

func ApplyRuleSetDefaults(rules *RuleSet) {
	if rules.Movies != nil && rules.Movies.UnknownCertification == "" {
		rules.Movies.UnknownCertification = string(enums.CertificationUnknownAllow)
	}
	if rules.Shows != nil && rules.Shows.UnknownCertification == "" {
		rules.Shows.UnknownCertification = string(enums.CertificationUnknownAllow)
	}
}

func (c *Config) RuleSetUsage(id string) int {
	count := 0
	for _, job := range c.Jobs.List {
		if job.RuleSetID == id {
			count++
		}
	}
	return count
}

// MigrateRuleSets upgrades global and embedded job filters deterministically in memory.
func (c *Config) MigrateRuleSets() {
	if _, ok := c.RuleSetByID(DefaultMoviesRuleSetID); !ok {
		filters := c.Filters.Movies
		c.RuleSets = append(c.RuleSets, RuleSet{ID: DefaultMoviesRuleSetID, Name: "Default Movies", Media: "movie", Revision: 1, Movies: &filters})
	}
	if _, ok := c.RuleSetByID(DefaultShowsRuleSetID); !ok {
		filters := c.Filters.Shows
		c.RuleSets = append(c.RuleSets, RuleSet{ID: DefaultShowsRuleSetID, Name: "Default Shows", Media: "show", Revision: 1, Shows: &filters})
	}
	for index := range c.RuleSets {
		ApplyRuleSetDefaults(&c.RuleSets[index])
	}
	for i := range c.Jobs.List {
		job := &c.Jobs.List[i]
		if job.RuleSetID != "" {
			continue
		}
		job.RuleSetID = defaultRuleSetID(job.MediaType)
		if job.UseCustomFilters {
			id := "migrated-" + job.ID
			if _, ok := c.RuleSetByID(id); !ok {
				rules := RuleSet{ID: id, Name: job.Name + " Rules", Media: job.MediaType, Revision: 1}
				if job.MediaType == "show" {
					filters := job.Filters.Shows
					rules.Shows = &filters
				} else {
					filters := job.Filters.Movies
					rules.Movies = &filters
				}
				c.RuleSets = append(c.RuleSets, rules)
			}
			job.RuleSetID = id
		}
	}
	c.TitleExceptions.BlockedMovieTMDBIDs = appendUniqueInts(c.TitleExceptions.BlockedMovieTMDBIDs, c.Filters.Movies.BlacklistedTMDBIds...)
	c.TitleExceptions.BlockedShowTVDBIDs = appendUniqueInts(c.TitleExceptions.BlockedShowTVDBIDs, c.Filters.Shows.BlacklistedTVDBIds...)
}

func appendUniqueInts(dst []int, values ...int) []int {
	seen := make(map[int]bool, len(dst)+len(values))
	for _, value := range dst {
		seen[value] = true
	}
	for _, value := range values {
		if value > 0 && !seen[value] {
			dst = append(dst, value)
			seen[value] = true
		}
	}
	return dst
}

// GetAllJobs returns a unified list of all jobs (dynamic list + legacy jobs converted to DynamicJob format)
func (c *Config) GetAllJobs() []DynamicJob {
	jobs := make([]DynamicJob, 0)

	// First, add all dynamic jobs from the list
	jobs = append(jobs, c.Jobs.List...)

	// Then, convert legacy jobs to DynamicJob format (only if not already in dynamic list)
	legacyJobs := c.getLegacyJobsAsDynamic()
	jobs = append(jobs, legacyJobs...)

	return jobs
}

// GetEnabledJobs returns only the enabled jobs from the unified list
func (c *Config) GetEnabledJobs() []DynamicJob {
	allJobs := c.GetAllJobs()
	enabled := make([]DynamicJob, 0)
	for _, job := range allJobs {
		if job.Enabled {
			enabled = append(enabled, job)
		}
	}
	return enabled
}

// GetDynamicJobByID returns a dynamic job by its ID
func (c *Config) GetDynamicJobByID(id string) *DynamicJob {
	for i := range c.Jobs.List {
		if c.Jobs.List[i].ID == id {
			return &c.Jobs.List[i]
		}
	}
	return nil
}

// AddDynamicJob adds a new job to the dynamic list
func (c *Config) AddDynamicJob(job DynamicJob) error {
	// Validate unique ID
	for _, existing := range c.Jobs.List {
		if existing.ID == job.ID {
			return fmt.Errorf("job with ID '%s' already exists", job.ID)
		}
	}

	// Set default source if not specified
	if job.Source == "" {
		job.Source = "trakt"
	}

	c.Jobs.List = append(c.Jobs.List, job)
	return nil
}

// UpdateDynamicJob updates an existing job in the dynamic list
func (c *Config) UpdateDynamicJob(job DynamicJob) error {
	if job.Source == "" {
		job.Source = "trakt"
	}
	for i, existing := range c.Jobs.List {
		if existing.ID == job.ID {
			c.Jobs.List[i] = job
			return nil
		}
	}
	return fmt.Errorf("job with ID '%s' not found", job.ID)
}

// DeleteDynamicJob removes a job from the dynamic list
func (c *Config) DeleteDynamicJob(id string) error {
	for i, existing := range c.Jobs.List {
		if existing.ID == id {
			c.Jobs.List = append(c.Jobs.List[:i], c.Jobs.List[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("job with ID '%s' not found", id)
}

// hasDynamicJobOfType checks if a dynamic job of given type and media already exists
func (c *Config) hasDynamicJobOfType(jobType, mediaType string) bool {
	for _, job := range c.Jobs.List {
		if job.Type == jobType && job.MediaType == mediaType {
			return true
		}
	}
	return false
}

// getLegacyJobsAsDynamic converts legacy fixed job structs to DynamicJob format
// Only includes legacy jobs that don't have a corresponding dynamic job of the same type
func (c *Config) getLegacyJobsAsDynamic() []DynamicJob {
	jobs := make([]DynamicJob, 0)

	// Trending Movies
	if c.Jobs.TrendingMovies.Enabled && !c.hasDynamicJobOfType("trending", "movie") {
		jobs = append(jobs, DynamicJob{
			ID:                  "legacy_trending_movies",
			Name:                "Trending Movies",
			Enabled:             c.Jobs.TrendingMovies.Enabled,
			Type:                "trending",
			Source:              "trakt",
			MediaType:           "movie",
			Limit:               c.Jobs.TrendingMovies.Limit,
			SyncInterval:        c.Jobs.TrendingMovies.SyncInterval,
			Mode:                c.Jobs.TrendingMovies.Mode,
			MinimumAvailability: c.Jobs.TrendingMovies.MinimumAvailability,
			Monitor:             c.Jobs.TrendingMovies.Monitor,
		})
	}

	// Trending Shows
	if c.Jobs.TrendingShows.Enabled && !c.hasDynamicJobOfType("trending", "show") {
		jobs = append(jobs, DynamicJob{
			ID:           "legacy_trending_shows",
			Name:         "Trending Shows",
			Enabled:      c.Jobs.TrendingShows.Enabled,
			Type:         "trending",
			Source:       "trakt",
			MediaType:    "show",
			Limit:        c.Jobs.TrendingShows.Limit,
			SyncInterval: c.Jobs.TrendingShows.SyncInterval,
			Mode:         c.Jobs.TrendingShows.Mode,
			Monitor:      c.Jobs.TrendingShows.Monitor,
		})
	}

	// Popular Movies
	if c.Jobs.PopularMovies.Enabled && !c.hasDynamicJobOfType("popular", "movie") {
		jobs = append(jobs, DynamicJob{
			ID:                  "legacy_popular_movies",
			Name:                "Popular Movies",
			Enabled:             c.Jobs.PopularMovies.Enabled,
			Type:                "popular",
			Source:              "trakt",
			MediaType:           "movie",
			Limit:               c.Jobs.PopularMovies.Limit,
			SyncInterval:        c.Jobs.PopularMovies.SyncInterval,
			Mode:                c.Jobs.PopularMovies.Mode,
			MinimumAvailability: c.Jobs.PopularMovies.MinimumAvailability,
			Monitor:             c.Jobs.PopularMovies.Monitor,
		})
	}

	// Popular Shows
	if c.Jobs.PopularShows.Enabled && !c.hasDynamicJobOfType("popular", "show") {
		jobs = append(jobs, DynamicJob{
			ID:           "legacy_popular_shows",
			Name:         "Popular Shows",
			Enabled:      c.Jobs.PopularShows.Enabled,
			Type:         "popular",
			Source:       "trakt",
			MediaType:    "show",
			Limit:        c.Jobs.PopularShows.Limit,
			SyncInterval: c.Jobs.PopularShows.SyncInterval,
			Mode:         c.Jobs.PopularShows.Mode,
			Monitor:      c.Jobs.PopularShows.Monitor,
		})
	}

	// Box Office
	if c.Jobs.BoxOffice.Enabled && !c.hasDynamicJobOfType("box_office", "movie") {
		jobs = append(jobs, DynamicJob{
			ID:                  "legacy_box_office",
			Name:                "Box Office",
			Enabled:             c.Jobs.BoxOffice.Enabled,
			Type:                "box_office",
			Source:              "trakt",
			MediaType:           "movie",
			Limit:               c.Jobs.BoxOffice.Limit,
			SyncInterval:        c.Jobs.BoxOffice.SyncInterval,
			Mode:                c.Jobs.BoxOffice.Mode,
			MinimumAvailability: c.Jobs.BoxOffice.MinimumAvailability,
			Monitor:             c.Jobs.BoxOffice.Monitor,
		})
	}

	// Favorited Movies
	if c.Jobs.FavoritedMovies.Enabled && !c.hasDynamicJobOfType("favorited", "movie") {
		jobs = append(jobs, DynamicJob{
			ID:                  "legacy_favorited_movies",
			Name:                "Favorited Movies",
			Enabled:             c.Jobs.FavoritedMovies.Enabled,
			Type:                "favorited",
			Source:              "trakt",
			MediaType:           "movie",
			Limit:               c.Jobs.FavoritedMovies.Limit,
			Period:              c.Jobs.FavoritedMovies.Period,
			SyncInterval:        c.Jobs.FavoritedMovies.SyncInterval,
			Mode:                c.Jobs.FavoritedMovies.Mode,
			MinimumAvailability: c.Jobs.FavoritedMovies.MinimumAvailability,
			Monitor:             c.Jobs.FavoritedMovies.Monitor,
		})
	}

	// Played Movies
	if c.Jobs.PlayedMovies.Enabled && !c.hasDynamicJobOfType("played", "movie") {
		jobs = append(jobs, DynamicJob{
			ID:                  "legacy_played_movies",
			Name:                "Played Movies",
			Enabled:             c.Jobs.PlayedMovies.Enabled,
			Type:                "played",
			Source:              "trakt",
			MediaType:           "movie",
			Limit:               c.Jobs.PlayedMovies.Limit,
			Period:              c.Jobs.PlayedMovies.Period,
			SyncInterval:        c.Jobs.PlayedMovies.SyncInterval,
			Mode:                c.Jobs.PlayedMovies.Mode,
			MinimumAvailability: c.Jobs.PlayedMovies.MinimumAvailability,
			Monitor:             c.Jobs.PlayedMovies.Monitor,
		})
	}

	// Watched Movies
	if c.Jobs.WatchedMovies.Enabled && !c.hasDynamicJobOfType("watched", "movie") {
		jobs = append(jobs, DynamicJob{
			ID:                  "legacy_watched_movies",
			Name:                "Watched Movies",
			Enabled:             c.Jobs.WatchedMovies.Enabled,
			Type:                "watched",
			Source:              "trakt",
			MediaType:           "movie",
			Limit:               c.Jobs.WatchedMovies.Limit,
			Period:              c.Jobs.WatchedMovies.Period,
			SyncInterval:        c.Jobs.WatchedMovies.SyncInterval,
			Mode:                c.Jobs.WatchedMovies.Mode,
			MinimumAvailability: c.Jobs.WatchedMovies.MinimumAvailability,
			Monitor:             c.Jobs.WatchedMovies.Monitor,
		})
	}

	// Collected Movies
	if c.Jobs.CollectedMovies.Enabled && !c.hasDynamicJobOfType("collected", "movie") {
		jobs = append(jobs, DynamicJob{
			ID:                  "legacy_collected_movies",
			Name:                "Collected Movies",
			Enabled:             c.Jobs.CollectedMovies.Enabled,
			Type:                "collected",
			Source:              "trakt",
			MediaType:           "movie",
			Limit:               c.Jobs.CollectedMovies.Limit,
			Period:              c.Jobs.CollectedMovies.Period,
			SyncInterval:        c.Jobs.CollectedMovies.SyncInterval,
			Mode:                c.Jobs.CollectedMovies.Mode,
			MinimumAvailability: c.Jobs.CollectedMovies.MinimumAvailability,
			Monitor:             c.Jobs.CollectedMovies.Monitor,
		})
	}

	// Anticipated Movies
	if c.Jobs.AnticipatedMovies.Enabled && !c.hasDynamicJobOfType("anticipated", "movie") {
		jobs = append(jobs, DynamicJob{
			ID:                  "legacy_anticipated_movies",
			Name:                "Anticipated Movies",
			Enabled:             c.Jobs.AnticipatedMovies.Enabled,
			Type:                "anticipated",
			Source:              "trakt",
			MediaType:           "movie",
			Limit:               c.Jobs.AnticipatedMovies.Limit,
			SyncInterval:        c.Jobs.AnticipatedMovies.SyncInterval,
			Mode:                c.Jobs.AnticipatedMovies.Mode,
			MinimumAvailability: c.Jobs.AnticipatedMovies.MinimumAvailability,
			Monitor:             c.Jobs.AnticipatedMovies.Monitor,
		})
	}

	// Favorited Shows
	if c.Jobs.FavoritedShows.Enabled && !c.hasDynamicJobOfType("favorited", "show") {
		jobs = append(jobs, DynamicJob{
			ID:           "legacy_favorited_shows",
			Name:         "Favorited Shows",
			Enabled:      c.Jobs.FavoritedShows.Enabled,
			Type:         "favorited",
			Source:       "trakt",
			MediaType:    "show",
			Limit:        c.Jobs.FavoritedShows.Limit,
			Period:       c.Jobs.FavoritedShows.Period,
			SyncInterval: c.Jobs.FavoritedShows.SyncInterval,
			Mode:         c.Jobs.FavoritedShows.Mode,
			Monitor:      c.Jobs.FavoritedShows.Monitor,
		})
	}

	// Played Shows
	if c.Jobs.PlayedShows.Enabled && !c.hasDynamicJobOfType("played", "show") {
		jobs = append(jobs, DynamicJob{
			ID:           "legacy_played_shows",
			Name:         "Played Shows",
			Enabled:      c.Jobs.PlayedShows.Enabled,
			Type:         "played",
			Source:       "trakt",
			MediaType:    "show",
			Limit:        c.Jobs.PlayedShows.Limit,
			Period:       c.Jobs.PlayedShows.Period,
			SyncInterval: c.Jobs.PlayedShows.SyncInterval,
			Mode:         c.Jobs.PlayedShows.Mode,
			Monitor:      c.Jobs.PlayedShows.Monitor,
		})
	}

	// Watched Shows
	if c.Jobs.WatchedShows.Enabled && !c.hasDynamicJobOfType("watched", "show") {
		jobs = append(jobs, DynamicJob{
			ID:           "legacy_watched_shows",
			Name:         "Watched Shows",
			Enabled:      c.Jobs.WatchedShows.Enabled,
			Type:         "watched",
			Source:       "trakt",
			MediaType:    "show",
			Limit:        c.Jobs.WatchedShows.Limit,
			Period:       c.Jobs.WatchedShows.Period,
			SyncInterval: c.Jobs.WatchedShows.SyncInterval,
			Mode:         c.Jobs.WatchedShows.Mode,
			Monitor:      c.Jobs.WatchedShows.Monitor,
		})
	}

	// Collected Shows
	if c.Jobs.CollectedShows.Enabled && !c.hasDynamicJobOfType("collected", "show") {
		jobs = append(jobs, DynamicJob{
			ID:           "legacy_collected_shows",
			Name:         "Collected Shows",
			Enabled:      c.Jobs.CollectedShows.Enabled,
			Type:         "collected",
			Source:       "trakt",
			MediaType:    "show",
			Limit:        c.Jobs.CollectedShows.Limit,
			Period:       c.Jobs.CollectedShows.Period,
			SyncInterval: c.Jobs.CollectedShows.SyncInterval,
			Mode:         c.Jobs.CollectedShows.Mode,
			Monitor:      c.Jobs.CollectedShows.Monitor,
		})
	}

	// Anticipated Shows
	if c.Jobs.AnticipatedShows.Enabled && !c.hasDynamicJobOfType("anticipated", "show") {
		jobs = append(jobs, DynamicJob{
			ID:           "legacy_anticipated_shows",
			Name:         "Anticipated Shows",
			Enabled:      c.Jobs.AnticipatedShows.Enabled,
			Type:         "anticipated",
			Source:       "trakt",
			MediaType:    "show",
			Limit:        c.Jobs.AnticipatedShows.Limit,
			SyncInterval: c.Jobs.AnticipatedShows.SyncInterval,
			Mode:         c.Jobs.AnticipatedShows.Mode,
			Monitor:      c.Jobs.AnticipatedShows.Monitor,
		})
	}

	// Smart Popular Movies
	if c.Jobs.SmartPopularMovies.Enabled && !c.hasDynamicJobOfType("smart_popular", "movie") {
		jobs = append(jobs, DynamicJob{
			ID:                  "legacy_smart_popular_movies",
			Name:                "Smart Popular Movies",
			Enabled:             c.Jobs.SmartPopularMovies.Enabled,
			Type:                "smart_popular",
			Source:              "trakt",
			MediaType:           "movie",
			Limit:               c.Jobs.SmartPopularMovies.Limit,
			SyncInterval:        c.Jobs.SmartPopularMovies.SyncInterval,
			Mode:                c.Jobs.SmartPopularMovies.Mode,
			MinimumAvailability: c.Jobs.SmartPopularMovies.MinimumAvailability,
			Monitor:             c.Jobs.SmartPopularMovies.Monitor,
			BaseMinRating:       c.Jobs.SmartPopularMovies.BaseMinRating,
			AdjustmentFactor:    c.Jobs.SmartPopularMovies.AdjustmentFactor,
		})
	}

	// Smart Popular Shows
	if c.Jobs.SmartPopularShows.Enabled && !c.hasDynamicJobOfType("smart_popular", "show") {
		jobs = append(jobs, DynamicJob{
			ID:               "legacy_smart_popular_shows",
			Name:             "Smart Popular Shows",
			Enabled:          c.Jobs.SmartPopularShows.Enabled,
			Type:             "smart_popular",
			Source:           "trakt",
			MediaType:        "show",
			Limit:            c.Jobs.SmartPopularShows.Limit,
			SyncInterval:     c.Jobs.SmartPopularShows.SyncInterval,
			Mode:             c.Jobs.SmartPopularShows.Mode,
			Monitor:          c.Jobs.SmartPopularShows.Monitor,
			BaseMinRating:    c.Jobs.SmartPopularShows.BaseMinRating,
			AdjustmentFactor: c.Jobs.SmartPopularShows.AdjustmentFactor,
		})
	}

	return jobs
}

// MigrateLegacyJobs converts all enabled legacy jobs to dynamic jobs and disables the legacy entries
func (c *Config) MigrateLegacyJobs() ([]DynamicJob, error) {
	migratedJobs := make([]DynamicJob, 0)
	originalJobs := append([]DynamicJob(nil), c.Jobs.List...)

	// Get legacy jobs as dynamic format
	legacyJobs := c.getLegacyJobsAsDynamic()

	for _, legacyJob := range legacyJobs {
		// Create a new ID without the legacy prefix
		newID := strings.TrimPrefix(legacyJob.ID, "legacy_")
		legacyJob.ID = newID
		legacyJob.RuleSetID = defaultRuleSetID(legacyJob.MediaType)

		// Add to dynamic list
		if err := c.AddDynamicJob(legacyJob); err != nil {
			c.Jobs.List = originalJobs
			return nil, fmt.Errorf("failed to migrate %q: %w", legacyJob.Name, err)
		}
		migratedJobs = append(migratedJobs, legacyJob)
	}

	// Disable all legacy jobs
	c.Jobs.TrendingMovies.Enabled = false
	c.Jobs.TrendingShows.Enabled = false
	c.Jobs.PopularMovies.Enabled = false
	c.Jobs.PopularShows.Enabled = false
	c.Jobs.BoxOffice.Enabled = false
	c.Jobs.FavoritedMovies.Enabled = false
	c.Jobs.PlayedMovies.Enabled = false
	c.Jobs.WatchedMovies.Enabled = false
	c.Jobs.CollectedMovies.Enabled = false
	c.Jobs.AnticipatedMovies.Enabled = false
	c.Jobs.FavoritedShows.Enabled = false
	c.Jobs.PlayedShows.Enabled = false
	c.Jobs.WatchedShows.Enabled = false
	c.Jobs.CollectedShows.Enabled = false
	c.Jobs.AnticipatedShows.Enabled = false
	c.Jobs.SmartPopularMovies.Enabled = false
	c.Jobs.SmartPopularShows.Enabled = false

	return migratedJobs, nil
}
