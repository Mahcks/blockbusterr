package jobs

import (
	"encoding/json"
	"strings"
	"time"
)

// FilterCheck represents a single filter evaluation result
type FilterCheck struct {
	Name    string `json:"name"`    // Filter name (e.g., "min_rating", "blacklisted_genres")
	Passed  bool   `json:"passed"`  // Whether this check passed
	Message string `json:"message"` // Human-readable explanation
}

// ContentDecision represents the complete decision for a single piece of content
type ContentDecision struct {
	// Basic Info
	Title     string `json:"title"`
	Year      int    `json:"year"`
	MediaType string `json:"media_type"` // "movie" or "show"
	TMDBID    int    `json:"tmdb_id"`
	TVDBID    int    `json:"tvdb_id,omitempty"`
	IMDBID    string `json:"imdb_id,omitempty"`
	PosterURL string `json:"poster_url,omitempty"`

	// Metrics
	Rating float64 `json:"rating"` // Trakt rating
	Votes  int     `json:"votes"`  // Vote count
	Score  float64 `json:"score"`  // Calculated score (0-1)
	Rank   int     `json:"rank"`   // Rank within this job's results

	// Decision Outcome
	PassedFilters bool   `json:"passed_filters"` // Whether it passed all filters
	Action        string `json:"action"`         // "added", "requested", "skipped", "rejected", "failed"
	ActionReason  string `json:"action_reason"`  // Why the action was taken

	// Filter Details
	FilterChecks []FilterCheck `json:"filter_checks"` // Detailed filter evaluation

	// Timing
	EvaluatedAt time.Time `json:"evaluated_at"`
}

// JobRunDecisions tracks all decisions for a single job run
type JobRunDecisions struct {
	JobName   string    `json:"job_name"`
	RunTime   time.Time `json:"run_time"`
	Completed bool      `json:"completed"`

	// Summary Stats
	TotalFound    int `json:"total_found"`
	PassedFilters int `json:"passed_filters"`
	Added         int `json:"added"`
	Requested     int `json:"requested"`
	Skipped       int `json:"skipped"`
	Rejected      int `json:"rejected"`
	Failed        int `json:"failed"`

	// All Decisions
	Decisions []ContentDecision `json:"decisions"`
}

// ToJSON serializes filter checks to JSON string for database storage
func FilterChecksToJSON(checks []FilterCheck) string {
	if len(checks) == 0 {
		return ""
	}

	data, err := json.Marshal(checks)
	if err != nil {
		return ""
	}

	return string(data)
}

// FromJSON deserializes filter checks from JSON string
func FilterChecksFromJSON(data string) []FilterCheck {
	if data == "" {
		return nil
	}

	var checks []FilterCheck
	if err := json.Unmarshal([]byte(data), &checks); err != nil {
		return nil
	}

	return checks
}

// HyphenToUnderscore converts hyphenated job names to underscore format
func HyphenToUnderscore(s string) string {
	return strings.ReplaceAll(s, "-", "_")
}
