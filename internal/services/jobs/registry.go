package jobs

// JobTypeDefinition describes a job type that can be instantiated as a DynamicJob
type JobTypeDefinition struct {
	Type           string   `json:"type"`            // Internal type identifier
	Name           string   `json:"name"`            // Display name
	Description    string   `json:"description"`     // UI description
	Source         string   `json:"source"`          // Data source: "trakt", "tmdb", etc.
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
	"trending": {
		Type:           "trending",
		Name:           "Trending",
		Description:    "Currently being watched and talked about",
		Source:         "trakt",
		SupportedMedia: []string{"movie", "show"},
		RequiresPeriod: false,
		IsSmartJob:     false,
		DefaultLimit:   50,
		MaxLimit:       1000,
	},
	"popular": {
		Type:           "popular",
		Name:           "Popular",
		Description:    "Most popular content on Trakt",
		Source:         "trakt",
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
		SupportedMedia: []string{"movie", "show"},
		RequiresPeriod: false,
		IsSmartJob:     true,
		DefaultLimit:   50,
		MaxLimit:       1000,
	},
}

// JobTemplate represents a pre-configured job template for quick setup
type JobTemplate struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Type        string  `json:"type"`
	MediaType   string  `json:"media"`
	Limit       int     `json:"limit"`
	Period      string  `json:"period,omitempty"`
	Category    string  `json:"category"` // "Movies" or "TV Shows"
}

// JobTemplates contains pre-defined job configurations for common use cases
var JobTemplates = []JobTemplate{
	// Movie Templates
	{
		Name:        "Weekly Trending Movies",
		Description: "Top 10 trending movies this week",
		Type:        "trending",
		MediaType:   "movie",
		Limit:       10,
		Category:    "Movies",
	},
	{
		Name:        "Monthly Popular Movies",
		Description: "Top 20 popular movies this month",
		Type:        "popular",
		MediaType:   "movie",
		Limit:       20,
		Category:    "Movies",
	},
	{
		Name:        "Weekly Watched Movies",
		Description: "Most watched movies this week",
		Type:        "watched",
		MediaType:   "movie",
		Limit:       20,
		Period:      "weekly",
		Category:    "Movies",
	},
	{
		Name:        "Monthly Watched Movies",
		Description: "Most watched movies this month",
		Type:        "watched",
		MediaType:   "movie",
		Limit:       30,
		Period:      "monthly",
		Category:    "Movies",
	},
	{
		Name:        "All-Time Top 50 Movies",
		Description: "Top 50 most watched movies of all time",
		Type:        "watched",
		MediaType:   "movie",
		Limit:       50,
		Period:      "all",
		Category:    "Movies",
	},
	{
		Name:        "Box Office Hits",
		Description: "Top 10 box office movies",
		Type:        "box_office",
		MediaType:   "movie",
		Limit:       10,
		Category:    "Movies",
	},
	{
		Name:        "Anticipated Movies",
		Description: "Top 25 most anticipated upcoming movies",
		Type:        "anticipated",
		MediaType:   "movie",
		Limit:       25,
		Category:    "Movies",
	},
	{
		Name:        "Smart Popular Movies",
		Description: "Popular movies with adaptive quality filtering",
		Type:        "smart_popular",
		MediaType:   "movie",
		Limit:       50,
		Category:    "Movies",
	},

	// TV Show Templates
	{
		Name:        "Weekly Trending Shows",
		Description: "Top 10 trending shows this week",
		Type:        "trending",
		MediaType:   "show",
		Limit:       10,
		Category:    "TV Shows",
	},
	{
		Name:        "Monthly Popular Shows",
		Description: "Top 20 popular shows this month",
		Type:        "popular",
		MediaType:   "show",
		Limit:       20,
		Category:    "TV Shows",
	},
	{
		Name:        "Weekly Watched Shows",
		Description: "Most watched shows this week",
		Type:        "watched",
		MediaType:   "show",
		Limit:       20,
		Period:      "weekly",
		Category:    "TV Shows",
	},
	{
		Name:        "All-Time Top 50 Shows",
		Description: "Top 50 most watched shows of all time",
		Type:        "watched",
		MediaType:   "show",
		Limit:       50,
		Period:      "all",
		Category:    "TV Shows",
	},
	{
		Name:        "Anticipated Shows",
		Description: "Top 25 most anticipated upcoming shows",
		Type:        "anticipated",
		MediaType:   "show",
		Limit:       25,
		Category:    "TV Shows",
	},
	{
		Name:        "Smart Popular Shows",
		Description: "Popular shows with adaptive quality filtering",
		Type:        "smart_popular",
		MediaType:   "show",
		Limit:       50,
		Category:    "TV Shows",
	},
}

// GetJobTypeDefinition returns the definition for a job type
func GetJobTypeDefinition(jobType string) (JobTypeDefinition, bool) {
	def, ok := JobTypeRegistry[jobType]
	return def, ok
}

// SupportsMediaType checks if a job type supports a given media type
func SupportsMediaType(jobType, mediaType string) bool {
	def, ok := JobTypeRegistry[jobType]
	if !ok {
		return false
	}
	for _, m := range def.SupportedMedia {
		if m == mediaType {
			return true
		}
	}
	return false
}

// GetAllJobTypes returns all job type definitions
func GetAllJobTypes() map[string]JobTypeDefinition {
	return JobTypeRegistry
}

// GetAllTemplates returns all job templates
func GetAllTemplates() []JobTemplate {
	return JobTemplates
}
