package jobs

import (
	"slices"
	"time"

	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/pkg/enums"
)

// JobTypeDefinition describes a job type that can be instantiated as a DynamicJob
type JobTypeDefinition struct {
	Type           string   `json:"type"`            // Internal type identifier
	Name           string   `json:"name"`            // Display name
	Description    string   `json:"description"`     // UI description
	Source         string   `json:"source"`          // Data source: "trakt", "tmdb", etc.
	Sources        []string `json:"sources"`         // Supported discovery sources.
	KnownSources   []string `json:"known_sources"`   // Sources shown disabled until their adapter is available.
	SupportedMedia []string `json:"supported_media"` // ["movie"], ["show"], or ["movie", "show"]
	RequiresPeriod bool     `json:"requires_period"` // Whether this job type uses period parameter
	IsSmartJob     bool     `json:"is_smart_job"`    // Whether this is a smart job with adaptive filters
	DefaultLimit   int      `json:"default_limit"`   // Default limit value
	MaxLimit       int      `json:"max_limit"`       // Maximum allowed limit
}

// JobTypeRegistry holds all available job type definitions
// Note: MaxLimit is set high (1000) for most types since Trakt supports pagination.
// Box Office is limited to 10 as that's all Trakt returns for that endpoint.
var JobTypeRegistry = map[string]JobTypeDefinition{
	string(enums.JobTypeList): {
		Type:           string(enums.JobTypeList),
		Name:           "List or Watchlist",
		Description:    "Discover content from a provider list or personal watchlist",
		Sources:        []string{"trakt", "tmdb", "letterboxd", "mdblist"},
		KnownSources:   []string{"trakt", "tmdb", "letterboxd", "mdblist"},
		SupportedMedia: []string{"movie", "show"},
		DefaultLimit:   100,
		MaxLimit:       1000,
	},
	"trending": {
		Type:           "trending",
		Name:           "Trending",
		Description:    "Currently being watched and talked about",
		Source:         "trakt",
		Sources:        []string{"trakt", "tmdb", "simkl"},
		SupportedMedia: []string{"movie", "show"},
		RequiresPeriod: false,
		IsSmartJob:     false,
		DefaultLimit:   50,
		MaxLimit:       1000,
	},
	"popular": {
		Type:           "popular",
		Name:           "Popular",
		Description:    "Most popular content from the selected source",
		Source:         "trakt",
		Sources:        []string{"trakt", "tmdb", "simkl"},
		SupportedMedia: []string{"movie", "show"},
		RequiresPeriod: false,
		IsSmartJob:     false,
		DefaultLimit:   50,
		MaxLimit:       1000,
	},
	"watched": {
		Type:           "watched",
		Name:           "Most Watched",
		Description:    "Most watched content over a time period",
		Source:         "trakt",
		Sources:        []string{"trakt", "simkl"},
		SupportedMedia: []string{"movie", "show"},
		RequiresPeriod: true,
		IsSmartJob:     false,
		DefaultLimit:   50,
		MaxLimit:       1000,
	},
	"collected": {
		Type:           "collected",
		Name:           "Most Collected",
		Description:    "Most collected content over a time period",
		Source:         "trakt",
		Sources:        []string{"trakt"},
		SupportedMedia: []string{"movie", "show"},
		RequiresPeriod: true,
		IsSmartJob:     false,
		DefaultLimit:   50,
		MaxLimit:       1000,
	},
	"favorited": {
		Type:           "favorited",
		Name:           "Most Favorited",
		Description:    "Most favorited content over a time period",
		Source:         "trakt",
		Sources:        []string{"trakt"},
		SupportedMedia: []string{"movie", "show"},
		RequiresPeriod: true,
		IsSmartJob:     false,
		DefaultLimit:   50,
		MaxLimit:       1000,
	},
	"played": {
		Type:           "played",
		Name:           "Most Played",
		Description:    "Most played content over a time period",
		Source:         "trakt",
		Sources:        []string{"trakt"},
		SupportedMedia: []string{"movie", "show"},
		RequiresPeriod: true,
		IsSmartJob:     false,
		DefaultLimit:   50,
		MaxLimit:       1000,
	},
	"anticipated": {
		Type:           "anticipated",
		Name:           "Anticipated",
		Description:    "Most anticipated upcoming content",
		Source:         "trakt",
		Sources:        []string{"trakt"},
		SupportedMedia: []string{"movie", "show"},
		RequiresPeriod: false,
		IsSmartJob:     false,
		DefaultLimit:   50,
		MaxLimit:       1000,
	},
	"box_office": {
		Type:           "box_office",
		Name:           "Box Office",
		Description:    "Top weekend box office movies (Trakt returns max 10)",
		Source:         "trakt",
		Sources:        []string{"trakt"},
		SupportedMedia: []string{"movie"},
		RequiresPeriod: false,
		IsSmartJob:     false,
		DefaultLimit:   10,
		MaxLimit:       10,
	},
	"smart_popular": {
		Type:           "smart_popular",
		Name:           "Smart Popular",
		Description:    "Popular content with adaptive rating thresholds",
		Source:         "trakt",
		Sources:        []string{"trakt", "tmdb", "simkl"},
		SupportedMedia: []string{"movie", "show"},
		RequiresPeriod: false,
		IsSmartJob:     true,
		DefaultLimit:   50,
		MaxLimit:       1000,
	},
}

// JobTemplate represents a pre-configured job template for quick setup
type JobTemplate struct {
	ID            string               `json:"id"`
	Version       int                  `json:"version"`
	Name          string               `json:"name"`
	Description   string               `json:"description"`
	Type          string               `json:"type"`
	MediaType     string               `json:"media"`
	Source        string               `json:"source"`
	Limit         int                  `json:"limit"`
	DeliveryLimit int                  `json:"delivery_limit"`
	Period        string               `json:"period,omitempty"`
	SyncInterval  string               `json:"sync_interval"`
	Mode          string               `json:"mode,omitempty"`
	Category      string               `json:"category"` // "Movies" or "TV Shows"
	RuleSetName   string               `json:"rule_set_name"`
	DefaultRules  bool                 `json:"default_rules,omitempty"`
	List          *config.ListLocator  `json:"list,omitempty"`
	Movies        *config.MovieFilters `json:"-"`
	Shows         *config.ShowFilters  `json:"-"`
}

type AvailableJobTemplate struct {
	JobTemplate
	Ready   bool   `json:"ready"`
	Missing string `json:"missing,omitempty"`
}

// JobTemplates contains pre-defined job configurations for common use cases
var JobTemplates = []JobTemplate{
	{ID: "balanced-trending-movies", Version: 1, Name: "Balanced Trending", Description: "A measured feed of current movies with baseline quality checks.", Type: "trending", MediaType: "movie", Source: "tmdb", Limit: 50, DeliveryLimit: 5, SyncInterval: "24h", Category: "Movies", RuleSetName: "Balanced Trending Movies", Movies: &config.MovieFilters{MinRating: 6.5, MinVotes: 250}},
	{ID: "new-well-rated-movies", Version: 1, Name: "New & Well Rated", Description: "Popular recent movies with stronger rating and vote requirements.", Type: "popular", MediaType: "movie", Source: "tmdb", Limit: 100, DeliveryLimit: 5, SyncInterval: "24h", Category: "Movies", RuleSetName: "New & Well Rated Movies", Movies: &config.MovieFilters{BlacklistedMinYear: time.Now().Year() - 1, MinRating: 7, MinVotes: 500}},
	{ID: "anticipated-approval", Version: 1, Name: "Anticipated With Approval", Description: "Upcoming movies sent through Jellyseerr or Seerr for approval.", Type: "anticipated", MediaType: "movie", Source: "trakt", Limit: 50, DeliveryLimit: 5, SyncInterval: "24h", Mode: "jellyseerr", Category: "Movies", RuleSetName: "Anticipated Movies", Movies: &config.MovieFilters{MinVotes: 100}},
	{ID: "documentary-discovery", Version: 1, Name: "Documentary Discovery", Description: "Well-rated documentary movies from TMDB.", Type: "popular", MediaType: "movie", Source: "tmdb", Limit: 100, DeliveryLimit: 3, SyncInterval: "168h", Category: "Movies", RuleSetName: "Documentary Movies", Movies: &config.MovieFilters{RequiredGenres: []string{"Documentary"}, MinRating: 6.5, MinVotes: 100}},
	{ID: "science-fiction-discovery", Version: 1, Name: "Science-Fiction Discovery", Description: "Quality science-fiction movies with enough audience signal.", Type: "popular", MediaType: "movie", Source: "tmdb", Limit: 100, DeliveryLimit: 5, SyncInterval: "168h", Category: "Movies", RuleSetName: "Science-Fiction Movies", Movies: &config.MovieFilters{RequiredGenres: []string{"Science Fiction"}, MinRating: 6.5, MinVotes: 250}},
	{ID: "family-movies", Version: 1, Name: "Family Movies", Description: "US G and PG family movies with unknown ratings rejected.", Type: "popular", MediaType: "movie", Source: "tmdb", Limit: 100, DeliveryLimit: 5, SyncInterval: "168h", Category: "Movies", RuleSetName: "Family Movies", Movies: &config.MovieFilters{RequiredGenres: []string{"Family"}, CertificationCountry: "US", AllowedCertifications: []string{"G", "PG"}, UnknownCertification: string(enums.CertificationUnknownReject)}},
	{ID: "all-time-classics", Version: 1, Name: "All-Time Classics", Description: "A small weekly selection from Trakt's all-time most watched movies.", Type: "watched", MediaType: "movie", Source: "trakt", Limit: 100, DeliveryLimit: 3, Period: "all", SyncInterval: "168h", Category: "Movies", RuleSetName: "All-Time Classics", Movies: &config.MovieFilters{BlacklistedMaxYear: time.Now().Year() - 10, MinRating: 7.5, MinVotes: 1000}},
	{ID: "personal-watchlist", Version: 1, Name: "Personal Watchlist", Description: "Sync your connected TMDB movie watchlist.", Type: string(enums.JobTypeList), MediaType: "movie", Source: "tmdb", Limit: 250, DeliveryLimit: 10, SyncInterval: "2h", Category: "Movies", RuleSetName: "Default Movies", DefaultRules: true, List: &config.ListLocator{Kind: string(enums.ListKindWatchlist), Ordering: "source"}},
	{ID: "balanced-trending-shows", Version: 1, Name: "Balanced Trending", Description: "Current TV with baseline quality checks and a conservative delivery cap.", Type: "trending", MediaType: "show", Source: "tmdb", Limit: 50, DeliveryLimit: 5, SyncInterval: "24h", Category: "TV Shows", RuleSetName: "Balanced Trending Shows", Shows: &config.ShowFilters{MinRating: 6.5, MinVotes: 100}},
	{ID: "reality-tv-discovery", Version: 1, Name: "Reality TV Discovery", Description: "Popular reality shows from TMDB.", Type: "popular", MediaType: "show", Source: "tmdb", Limit: 100, DeliveryLimit: 5, SyncInterval: "168h", Category: "TV Shows", RuleSetName: "Reality TV", Shows: &config.ShowFilters{RequiredGenres: []string{"Reality"}, MinRating: 6}},
	{ID: "current-tv", Version: 1, Name: "Current TV", Description: "Trending shows first aired within the last two years.", Type: "trending", MediaType: "show", Source: "tmdb", Limit: 100, DeliveryLimit: 5, SyncInterval: "24h", Category: "TV Shows", RuleSetName: "Current TV", Shows: &config.ShowFilters{BlacklistedMinYear: time.Now().Year() - 1, MinRating: 6.5}},
	{ID: "anime-discovery", Version: 1, Name: "Anime Discovery", Description: "Japanese animation that must map cleanly into Sonarr.", Type: "popular", MediaType: "show", Source: "tmdb", Limit: 100, DeliveryLimit: 5, SyncInterval: "168h", Category: "TV Shows", RuleSetName: "Anime Discovery", Shows: &config.ShowFilters{AllowedCountries: []string{"JP"}, RequiredGenres: []string{"Animation"}, MinRating: 6.5}},
}

// GetJobTypeDefinition returns the definition for a job type
func GetJobTypeDefinition(jobType string) (JobTypeDefinition, bool) {
	def, ok := JobTypeRegistry[jobType]
	return def, ok
}

func SupportsSource(jobType, source string) bool {
	definition, ok := JobTypeRegistry[jobType]
	if !ok {
		return false
	}
	if source == "" {
		source = definition.Source
	}
	return slices.Contains(definition.Sources, source)
}

// SupportsMediaType checks if a job type supports a given media type
func SupportsMediaType(jobType, mediaType string) bool {
	def, ok := JobTypeRegistry[jobType]
	if !ok {
		return false
	}
	return slices.Contains(def.SupportedMedia, mediaType)
}

// GetAvailableJobTypes limits each definition to configured providers. Empty
// definitions are retained so existing jobs remain understandable in the UI.
func GetAvailableJobTypes(providers, listProviders []string) map[string]JobTypeDefinition {
	definitions := make(map[string]JobTypeDefinition, len(JobTypeRegistry))
	for jobType, definition := range JobTypeRegistry {
		if len(definition.KnownSources) == 0 {
			definition.KnownSources = slices.Clone(definition.Sources)
		}
		available := providers
		if jobType == string(enums.JobTypeList) {
			available = listProviders
		}
		definition.Sources = slices.DeleteFunc(slices.Clone(definition.Sources), func(source string) bool {
			return !slices.Contains(available, source)
		})
		definitions[jobType] = definition
	}
	return definitions
}

// GetAllTemplates returns all job templates
func GetAllTemplates(cfg *config.Config) []AvailableJobTemplate {
	result := make([]AvailableJobTemplate, 0, len(JobTemplates))
	for _, recipe := range JobTemplates {
		ready, missing := recipeReadiness(cfg, recipe)
		result = append(result, AvailableJobTemplate{JobTemplate: recipe, Ready: ready, Missing: missing})
	}
	return result
}

func GetTemplate(id string) (JobTemplate, bool) {
	for _, recipe := range JobTemplates {
		if recipe.ID == id {
			return recipe, true
		}
	}
	return JobTemplate{}, false
}

func recipeReadiness(cfg *config.Config, recipe JobTemplate) (bool, string) {
	job := config.DynamicJob{Type: recipe.Type, Source: recipe.Source}
	if !IsJobSourceConfigured(cfg, job) {
		return false, "Connect " + recipe.Source
	}
	if recipe.Type == string(enums.JobTypeList) && recipe.List != nil && enums.ListKind(recipe.List.Kind) == enums.ListKindWatchlist && (cfg.TMDB.SessionID == "" || cfg.TMDB.AccountID == 0) {
		return false, "Connect your TMDB account"
	}
	mode := DetermineMode(recipe.Mode, cfg.Jobs.Mode)
	if mode == "jellyseerr" && (cfg.Jellyseerr.URL == "" || cfg.Jellyseerr.APIKey == "") {
		return false, "Connect Jellyseerr or Seerr"
	}
	if mode == "direct" && recipe.MediaType == "movie" && (cfg.Radarr.URL == "" || cfg.Radarr.APIKey == "") {
		return false, "Connect Radarr"
	}
	if mode == "direct" && recipe.MediaType == "show" && (cfg.Sonarr.URL == "" || cfg.Sonarr.APIKey == "") {
		return false, "Connect Sonarr"
	}
	return true, ""
}
