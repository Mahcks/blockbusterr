package filters

import (
	"fmt"
	"slices"
	"strings"

	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/integrations"
)

// FilterCheck represents a single filter check result
type FilterCheck struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Message string `json:"message"`
}

// FilterResult contains all filter checks for a piece of content
type FilterResult struct {
	Passed bool          `json:"passed"`
	Reason string        `json:"reason,omitempty"` // Overall reason if failed
	Checks []FilterCheck `json:"checks"`
}

// Explain turns an evaluation into the same concise, user-facing decision
// sentence used by previews and persisted Activity Entries.
func Explain(result FilterResult) string {
	if result.Passed {
		for _, check := range result.Checks {
			if check.Name == "Title exceptions" && check.Passed {
				return "Accepted: Universal title exception"
			}
			if check.Name == "Allow override" && check.Passed {
				return "Accepted: " + check.Message
			}
		}
		return "Accepted: Passed all configured rules"
	}
	for _, check := range slices.Backward(result.Checks) {
		if !check.Passed && check.Message != "" {
			return "Rejected: " + check.Message
		}
	}
	if result.Reason != "" {
		return "Rejected: " + result.Reason
	}
	return "Rejected: Did not pass the configured rules"
}

func MoviePassesRules(movie integrations.Movie, rules config.MovieFilters, exceptions config.TitleExceptions) FilterResult {
	if slices.Contains(exceptions.BlockedMovieTMDBIDs, movie.IDs.TMDB) {
		return FilterResult{Passed: false, Reason: "blocked title", Checks: []FilterCheck{{Name: "Title exceptions", Passed: false, Message: "Title is globally blocked"}}}
	}
	if slices.Contains(exceptions.AllowedMovieTMDBIDs, movie.IDs.TMDB) {
		return FilterResult{Passed: true, Checks: []FilterCheck{{Name: "Title exceptions", Passed: true, Message: "Title is globally allowed"}}}
	}
	if blocked := movieHardBlock(movie, rules); blocked != nil {
		return *blocked
	}
	if allowed := movieAllowOverride(movie, rules); allowed != nil {
		return *allowed
	}
	return MoviePassesFiltersDetailed(movie, rules)
}

func ShowPassesRules(show integrations.Show, rules config.ShowFilters, exceptions config.TitleExceptions) FilterResult {
	if slices.Contains(exceptions.BlockedShowTVDBIDs, show.IDs.TVDB) {
		return FilterResult{Passed: false, Reason: "blocked title", Checks: []FilterCheck{{Name: "Title exceptions", Passed: false, Message: "Title is globally blocked"}}}
	}
	if slices.Contains(exceptions.AllowedShowTVDBIDs, show.IDs.TVDB) {
		return FilterResult{Passed: true, Checks: []FilterCheck{{Name: "Title exceptions", Passed: true, Message: "Title is globally allowed"}}}
	}
	if blocked := showHardBlock(show, rules); blocked != nil {
		return *blocked
	}
	if allowed := showAllowOverride(show, rules); allowed != nil {
		return *allowed
	}
	return ShowPassesFiltersDetailed(show, rules)
}

func movieHardBlock(movie integrations.Movie, rules config.MovieFilters) *FilterResult {
	if slices.Contains(rules.BlacklistedTMDBIds, movie.IDs.TMDB) {
		return failed("Blocked title", fmt.Sprintf("TMDB ID %d is blocked", movie.IDs.TMDB))
	}
	if containsIgnoreCase(rules.BlacklistedCountries, movie.Country) {
		return failed("Blocked country", fmt.Sprintf("Blocked country matched: %s", movie.Country))
	}
	if containsIgnoreCase(rules.BlacklistedLanguages, movie.Language) {
		return failed("Blocked language", fmt.Sprintf("Blocked language matched: %s", movie.Language))
	}
	if match := firstMatch(rules.BlacklistedGenres, movie.Genres); match != "" {
		return failed("Blocked genre", fmt.Sprintf("Blocked genre matched: %s", match))
	}
	if match := matchingKeyword(rules.BlacklistedKeywords, movie.Title); match != "" {
		return failed("Blocked keyword", fmt.Sprintf("Blocked title keyword matched: %s", match))
	}
	return nil
}

func showHardBlock(show integrations.Show, rules config.ShowFilters) *FilterResult {
	if slices.Contains(rules.BlacklistedTVDBIds, show.IDs.TVDB) {
		return failed("Blocked title", fmt.Sprintf("TVDB ID %d is blocked", show.IDs.TVDB))
	}
	if containsIgnoreCase(rules.BlacklistedCountries, show.Country) {
		return failed("Blocked country", fmt.Sprintf("Blocked country matched: %s", show.Country))
	}
	if containsIgnoreCase(rules.BlacklistedLanguages, show.Language) {
		return failed("Blocked language", fmt.Sprintf("Blocked language matched: %s", show.Language))
	}
	if match := firstMatch(rules.BlacklistedGenres, show.Genres); match != "" {
		return failed("Blocked genre", fmt.Sprintf("Blocked genre matched: %s", match))
	}
	if containsIgnoreCase(rules.BlacklistedNetworks, show.Network) {
		return failed("Blocked network", fmt.Sprintf("Blocked network matched: %s", show.Network))
	}
	if match := matchingKeyword(rules.BlacklistedKeywords, show.Title); match != "" {
		return failed("Blocked keyword", fmt.Sprintf("Blocked title keyword matched: %s", match))
	}
	return nil
}

func movieAllowOverride(movie integrations.Movie, rules config.MovieFilters) *FilterResult {
	return allowOverride(movie.Country, movie.Language, movie.Genres, movie.Title, "", movie.Rating, rules.AllowCountries, rules.AllowLanguages, rules.AllowGenres, rules.AllowKeywords, nil, rules.AllowMinRating)
}

func showAllowOverride(show integrations.Show, rules config.ShowFilters) *FilterResult {
	return allowOverride(show.Country, show.Language, show.Genres, show.Title, show.Network, show.Rating, rules.AllowCountries, rules.AllowLanguages, rules.AllowGenres, rules.AllowKeywords, rules.AllowNetworks, rules.AllowMinRating)
}

func allowOverride(country, language string, genres []string, title, network string, rating float64, countries, languages, allowedGenres, keywords, networks []string, minRating float64) *FilterResult {
	name, value := "", ""
	switch {
	case containsIgnoreCase(countries, country):
		name, value = "country", country
	case containsIgnoreCase(languages, language):
		name, value = "language", language
	case firstMatch(allowedGenres, genres) != "":
		name, value = "genre", firstMatch(allowedGenres, genres)
	case matchingKeyword(keywords, title) != "":
		name, value = "title keyword", matchingKeyword(keywords, title)
	case containsIgnoreCase(networks, network):
		name, value = "network", network
	case minRating > 0 && rating >= minRating:
		name, value = "minimum rating", fmt.Sprintf("%.1f", minRating)
	default:
		return nil
	}
	message := fmt.Sprintf("Allow override matched %s: %s", name, value)
	return &FilterResult{Passed: true, Reason: message, Checks: []FilterCheck{{Name: "Allow override", Passed: true, Message: message}}}
}

func failed(name, message string) *FilterResult {
	return &FilterResult{Passed: false, Reason: strings.ToLower(message), Checks: []FilterCheck{{Name: name, Passed: false, Message: message}}}
}

// MoviePassesFilters checks if a movie passes all configured filters
func MoviePassesFilters(movie integrations.Movie, filters config.MovieFilters) (bool, string) {
	result := MoviePassesFiltersDetailed(movie, filters)
	return result.Passed, result.Reason
}

// MoviePassesFiltersDetailed checks if a movie passes all configured filters and returns detailed results
func MoviePassesFiltersDetailed(movie integrations.Movie, filters config.MovieFilters) FilterResult {
	result := FilterResult{
		Passed: true,
		Checks: make([]FilterCheck, 0),
	}

	// Check TMDB ID blacklist
	if len(filters.BlacklistedTMDBIds) > 0 {
		if slices.Contains(filters.BlacklistedTMDBIds, movie.IDs.TMDB) {
			result.Checks = append(result.Checks, FilterCheck{
				Name:    "TMDB Blacklist",
				Passed:  false,
				Message: fmt.Sprintf("TMDB ID %d is blacklisted", movie.IDs.TMDB),
			})
			result.Passed = false
			result.Reason = "blacklisted TMDB ID"
			return result
		}
		result.Checks = append(result.Checks, FilterCheck{
			Name:    "TMDB Blacklist",
			Passed:  true,
			Message: "Not in TMDB blacklist",
		})
	}
	if containsIgnoreCase(filters.BlacklistedCountries, movie.Country) {
		return *failed("Blocked country", fmt.Sprintf("Blocked country matched: %s", movie.Country))
	}
	if containsIgnoreCase(filters.BlacklistedLanguages, movie.Language) {
		return *failed("Blocked language", fmt.Sprintf("Blocked language matched: %s", movie.Language))
	}

	// Check country filter
	if len(filters.AllowedCountries) > 0 && !slices.Contains(filters.AllowedCountries, "ignore") {
		if movie.Country == "" {
			return FilterResult{Passed: false, Reason: "country metadata unavailable", Checks: []FilterCheck{{Name: "Allowed Countries", Passed: false, Message: "Country metadata is unavailable"}}}
		}
		if movie.Country != "" {
			if !containsIgnoreCase(filters.AllowedCountries, movie.Country) {
				result.Checks = append(result.Checks, FilterCheck{
					Name:    "Allowed Countries",
					Passed:  false,
					Message: fmt.Sprintf("Country '%s' not in allowed list", movie.Country),
				})
				result.Passed = false
				result.Reason = "country not allowed: " + movie.Country
				return result
			}
			result.Checks = append(result.Checks, FilterCheck{
				Name:    "Allowed Countries",
				Passed:  true,
				Message: fmt.Sprintf("Country '%s' is allowed", movie.Country),
			})
		}
	}

	// Check language filter
	if len(filters.AllowedLanguages) > 0 && !slices.Contains(filters.AllowedLanguages, "ignore") {
		if movie.Language == "" {
			return FilterResult{Passed: false, Reason: "language metadata unavailable", Checks: []FilterCheck{{Name: "Allowed Languages", Passed: false, Message: "Language metadata is unavailable"}}}
		}
		if movie.Language != "" {
			if !containsIgnoreCase(filters.AllowedLanguages, movie.Language) {
				result.Checks = append(result.Checks, FilterCheck{
					Name:    "Allowed Languages",
					Passed:  false,
					Message: fmt.Sprintf("Language '%s' not in allowed list", movie.Language),
				})
				result.Passed = false
				result.Reason = "language not allowed: " + movie.Language
				return result
			}
			result.Checks = append(result.Checks, FilterCheck{
				Name:    "Allowed Languages",
				Passed:  true,
				Message: fmt.Sprintf("Language '%s' is allowed", movie.Language),
			})
		}
	}

	// Check genre blacklist
	if len(filters.BlacklistedGenres) > 0 && !slices.Contains(filters.BlacklistedGenres, "ignore") {
		for _, genre := range movie.Genres {
			if containsIgnoreCase(filters.BlacklistedGenres, genre) {
				result.Checks = append(result.Checks, FilterCheck{
					Name:    "blacklisted_genres",
					Passed:  false,
					Message: fmt.Sprintf("Genre '%s' is blacklisted", genre),
				})
				result.Passed = false
				result.Reason = "blacklisted genre: " + genre
				return result
			}
		}
		if len(movie.Genres) > 0 {
			result.Checks = append(result.Checks, FilterCheck{
				Name:    "Blacklisted Genres",
				Passed:  true,
				Message: "No blacklisted genres",
			})
		}
	}

	// Check title keywords blacklist
	if len(filters.BlacklistedKeywords) > 0 {
		titleLower := strings.ToLower(movie.Title)
		blocked := false
		var blockedKeyword string
		for _, keyword := range filters.BlacklistedKeywords {
			if strings.Contains(titleLower, strings.ToLower(keyword)) {
				blocked = true
				blockedKeyword = keyword
				break
			}
		}
		if blocked {
			result.Checks = append(result.Checks, FilterCheck{
				Name:    "Blacklisted Keywords",
				Passed:  false,
				Message: fmt.Sprintf("Title contains blacklisted keyword '%s'", blockedKeyword),
			})
			result.Passed = false
			result.Reason = "blacklisted keyword in title: " + blockedKeyword
			return result
		}
		result.Checks = append(result.Checks, FilterCheck{
			Name:    "Blacklisted Keywords",
			Passed:  true,
			Message: "No blacklisted keywords in title",
		})
	}

	required, failure := requiredChecks(movie.Genres, movie.Title, "", filters.RequiredGenres, filters.RequiredKeywords, nil)
	if failure != nil {
		return *failure
	}
	result.Checks = append(result.Checks, required...)

	// Check runtime filters
	if movie.Runtime > 0 {
		if filters.BlacklistedMinRuntime > 0 && movie.Runtime < filters.BlacklistedMinRuntime {
			result.Checks = append(result.Checks, FilterCheck{
				Name:    "Minimum Runtime",
				Passed:  false,
				Message: fmt.Sprintf("Runtime %d min below minimum %d min", movie.Runtime, filters.BlacklistedMinRuntime),
			})
			result.Passed = false
			result.Reason = "runtime too short"
			return result
		}
		if filters.BlacklistedMaxRuntime > 0 && movie.Runtime > filters.BlacklistedMaxRuntime {
			result.Checks = append(result.Checks, FilterCheck{
				Name:    "Maximum Runtime",
				Passed:  false,
				Message: fmt.Sprintf("Runtime %d min exceeds maximum %d min", movie.Runtime, filters.BlacklistedMaxRuntime),
			})
			result.Passed = false
			result.Reason = "runtime too long"
			return result
		}
		if filters.BlacklistedMinRuntime > 0 || filters.BlacklistedMaxRuntime > 0 {
			result.Checks = append(result.Checks, FilterCheck{
				Name:    "Runtime",
				Passed:  true,
				Message: fmt.Sprintf("Runtime %d min within allowed range", movie.Runtime),
			})
		}
	}

	// Check year filters
	if movie.Year > 0 {
		if filters.BlacklistedMinYear > 0 && movie.Year < filters.BlacklistedMinYear {
			result.Checks = append(result.Checks, FilterCheck{
				Name:    "Minimum Year",
				Passed:  false,
				Message: fmt.Sprintf("Year %d before minimum %d", movie.Year, filters.BlacklistedMinYear),
			})
			result.Passed = false
			result.Reason = "year too old"
			return result
		}
		if filters.BlacklistedMaxYear > 0 && movie.Year > filters.BlacklistedMaxYear {
			result.Checks = append(result.Checks, FilterCheck{
				Name:    "Maximum Year",
				Passed:  false,
				Message: fmt.Sprintf("Year %d after maximum %d", movie.Year, filters.BlacklistedMaxYear),
			})
			result.Passed = false
			result.Reason = "year too new"
			return result
		}
		if filters.BlacklistedMinYear > 0 || filters.BlacklistedMaxYear > 0 {
			result.Checks = append(result.Checks, FilterCheck{
				Name:    "Release Year",
				Passed:  true,
				Message: fmt.Sprintf("Year %d within allowed range", movie.Year),
			})
		}
	}

	// Check rating filter
	if filters.MinRating > 0 {
		if movie.Rating < filters.MinRating {
			result.Checks = append(result.Checks, FilterCheck{
				Name:    "Minimum Rating",
				Passed:  false,
				Message: fmt.Sprintf("Rating %.1f below minimum %.1f", movie.Rating, filters.MinRating),
			})
			result.Passed = false
			result.Reason = fmt.Sprintf("rating %.1f below minimum %.1f", movie.Rating, filters.MinRating)
			return result
		}
		result.Checks = append(result.Checks, FilterCheck{
			Name:    "Minimum Rating",
			Passed:  true,
			Message: fmt.Sprintf("Rating %.1f meets minimum %.1f", movie.Rating, filters.MinRating),
		})
	}

	// Check minimum votes filter
	if filters.MinVotes > 0 {
		if movie.Votes < filters.MinVotes {
			result.Checks = append(result.Checks, FilterCheck{
				Name:    "Minimum Votes",
				Passed:  false,
				Message: fmt.Sprintf("Votes %d below minimum %d", movie.Votes, filters.MinVotes),
			})
			result.Passed = false
			result.Reason = fmt.Sprintf("votes %d below minimum %d", movie.Votes, filters.MinVotes)
			return result
		}
		result.Checks = append(result.Checks, FilterCheck{
			Name:    "Minimum Votes",
			Passed:  true,
			Message: fmt.Sprintf("Votes %d meets minimum %d", movie.Votes, filters.MinVotes),
		})
	}

	return result
}

// ShowPassesFiltersDetailed checks if a show passes all configured filters with detailed tracking
func ShowPassesFiltersDetailed(show integrations.Show, filters config.ShowFilters) FilterResult {
	result := FilterResult{
		Passed: true,
		Checks: []FilterCheck{},
	}

	// Check TVDB ID blacklist
	if len(filters.BlacklistedTVDBIds) > 0 {
		for _, id := range filters.BlacklistedTVDBIds {
			if show.IDs.TVDB == id {
				result.Checks = append(result.Checks, FilterCheck{
					Name:    "TVDB Blacklist",
					Passed:  false,
					Message: fmt.Sprintf("TVDB ID %d is blacklisted", id),
				})
				result.Passed = false
				result.Reason = "blacklisted TVDB ID"
				return result
			}
		}
		result.Checks = append(result.Checks, FilterCheck{
			Name:    "TVDB Blacklist",
			Passed:  true,
			Message: "TVDB ID not in blacklist",
		})
	}
	if containsIgnoreCase(filters.BlacklistedCountries, show.Country) {
		return *failed("Blocked country", fmt.Sprintf("Blocked country matched: %s", show.Country))
	}
	if containsIgnoreCase(filters.BlacklistedLanguages, show.Language) {
		return *failed("Blocked language", fmt.Sprintf("Blocked language matched: %s", show.Language))
	}

	// Check country filter
	if len(filters.AllowedCountries) > 0 && !slices.Contains(filters.AllowedCountries, "ignore") {
		if show.Country == "" {
			return FilterResult{Passed: false, Reason: "country metadata unavailable", Checks: []FilterCheck{{Name: "Allowed Countries", Passed: false, Message: "Country metadata is unavailable"}}}
		}
		if show.Country != "" && !containsIgnoreCase(filters.AllowedCountries, show.Country) {
			result.Checks = append(result.Checks, FilterCheck{
				Name:    "Allowed Countries",
				Passed:  false,
				Message: fmt.Sprintf("Country '%s' not in allowed list", show.Country),
			})
			result.Passed = false
			result.Reason = "country not allowed: " + show.Country
			return result
		}
		result.Checks = append(result.Checks, FilterCheck{
			Name:    "Allowed Countries",
			Passed:  true,
			Message: fmt.Sprintf("Country '%s' is allowed", show.Country),
		})
	}

	// Check language filter
	if len(filters.AllowedLanguages) > 0 && !slices.Contains(filters.AllowedLanguages, "ignore") {
		if show.Language == "" {
			return FilterResult{Passed: false, Reason: "language metadata unavailable", Checks: []FilterCheck{{Name: "Allowed Languages", Passed: false, Message: "Language metadata is unavailable"}}}
		}
		if show.Language != "" && !containsIgnoreCase(filters.AllowedLanguages, show.Language) {
			result.Checks = append(result.Checks, FilterCheck{
				Name:    "Allowed Languages",
				Passed:  false,
				Message: fmt.Sprintf("Language '%s' not in allowed list", show.Language),
			})
			result.Passed = false
			result.Reason = "language not allowed: " + show.Language
			return result
		}
		result.Checks = append(result.Checks, FilterCheck{
			Name:    "Allowed Languages",
			Passed:  true,
			Message: fmt.Sprintf("Language '%s' is allowed", show.Language),
		})
	}

	// Check genre blacklist
	if len(filters.BlacklistedGenres) > 0 && !slices.Contains(filters.BlacklistedGenres, "ignore") {
		for _, genre := range show.Genres {
			if containsIgnoreCase(filters.BlacklistedGenres, genre) {
				result.Checks = append(result.Checks, FilterCheck{
					Name:    "blacklisted_genres",
					Passed:  false,
					Message: fmt.Sprintf("Genre '%s' is blacklisted", genre),
				})
				result.Passed = false
				result.Reason = "blacklisted genre: " + genre
				return result
			}
		}
		result.Checks = append(result.Checks, FilterCheck{
			Name:    "Blacklisted Genres",
			Passed:  true,
			Message: "No blacklisted genres found",
		})
	}

	// Check network blacklist
	if len(filters.BlacklistedNetworks) > 0 {
		if show.Network != "" && containsIgnoreCase(filters.BlacklistedNetworks, show.Network) {
			result.Checks = append(result.Checks, FilterCheck{
				Name:    "Blacklisted Networks",
				Passed:  false,
				Message: fmt.Sprintf("Network '%s' is blacklisted", show.Network),
			})
			result.Passed = false
			result.Reason = "blacklisted network: " + show.Network
			return result
		}
		result.Checks = append(result.Checks, FilterCheck{
			Name:    "Blacklisted Networks",
			Passed:  true,
			Message: fmt.Sprintf("Network '%s' is allowed", show.Network),
		})
	}

	// Check title keywords blacklist
	if len(filters.BlacklistedKeywords) > 0 {
		titleLower := strings.ToLower(show.Title)
		for _, keyword := range filters.BlacklistedKeywords {
			if strings.Contains(titleLower, strings.ToLower(keyword)) {
				result.Checks = append(result.Checks, FilterCheck{
					Name:    "Blacklisted Keywords",
					Passed:  false,
					Message: fmt.Sprintf("Title contains blacklisted keyword '%s'", keyword),
				})
				result.Passed = false
				result.Reason = "blacklisted keyword in title: " + keyword
				return result
			}
		}
		result.Checks = append(result.Checks, FilterCheck{
			Name:    "Blacklisted Keywords",
			Passed:  true,
			Message: "No blacklisted keywords in title",
		})
	}

	required, failure := requiredChecks(show.Genres, show.Title, show.Network, filters.RequiredGenres, filters.RequiredKeywords, filters.RequiredNetworks)
	if failure != nil {
		return *failure
	}
	result.Checks = append(result.Checks, required...)

	// Check runtime filters
	if show.Runtime > 0 {
		if filters.BlacklistedMinRuntime > 0 && show.Runtime < filters.BlacklistedMinRuntime {
			result.Checks = append(result.Checks, FilterCheck{
				Name:    "Minimum Runtime",
				Passed:  false,
				Message: fmt.Sprintf("Runtime %d min below minimum %d min", show.Runtime, filters.BlacklistedMinRuntime),
			})
			result.Passed = false
			result.Reason = "runtime too short"
			return result
		}
		if filters.BlacklistedMaxRuntime > 0 && show.Runtime > filters.BlacklistedMaxRuntime {
			result.Checks = append(result.Checks, FilterCheck{
				Name:    "Maximum Runtime",
				Passed:  false,
				Message: fmt.Sprintf("Runtime %d min exceeds maximum %d min", show.Runtime, filters.BlacklistedMaxRuntime),
			})
			result.Passed = false
			result.Reason = "runtime too long"
			return result
		}
		if filters.BlacklistedMinRuntime > 0 || filters.BlacklistedMaxRuntime > 0 {
			result.Checks = append(result.Checks, FilterCheck{
				Name:    "Runtime",
				Passed:  true,
				Message: fmt.Sprintf("Runtime %d min within allowed range", show.Runtime),
			})
		}
	}

	// Check year filters
	if show.Year > 0 {
		if filters.BlacklistedMinYear > 0 && show.Year < filters.BlacklistedMinYear {
			result.Checks = append(result.Checks, FilterCheck{
				Name:    "Minimum Year",
				Passed:  false,
				Message: fmt.Sprintf("Year %d below minimum %d", show.Year, filters.BlacklistedMinYear),
			})
			result.Passed = false
			result.Reason = "year too old"
			return result
		}
		if filters.BlacklistedMaxYear > 0 && show.Year > filters.BlacklistedMaxYear {
			result.Checks = append(result.Checks, FilterCheck{
				Name:    "Maximum Year",
				Passed:  false,
				Message: fmt.Sprintf("Year %d exceeds maximum %d", show.Year, filters.BlacklistedMaxYear),
			})
			result.Passed = false
			result.Reason = "year too new"
			return result
		}
		if filters.BlacklistedMinYear > 0 || filters.BlacklistedMaxYear > 0 {
			result.Checks = append(result.Checks, FilterCheck{
				Name:    "Release Year",
				Passed:  true,
				Message: fmt.Sprintf("Year %d within allowed range", show.Year),
			})
		}
	}

	// Check rating filter
	if filters.MinRating > 0 {
		if show.Rating < filters.MinRating {
			result.Checks = append(result.Checks, FilterCheck{
				Name:    "Minimum Rating",
				Passed:  false,
				Message: fmt.Sprintf("Rating %.1f below minimum %.1f", show.Rating, filters.MinRating),
			})
			result.Passed = false
			result.Reason = fmt.Sprintf("rating %.1f below minimum %.1f", show.Rating, filters.MinRating)
			return result
		}
		result.Checks = append(result.Checks, FilterCheck{
			Name:    "Minimum Rating",
			Passed:  true,
			Message: fmt.Sprintf("Rating %.1f meets minimum %.1f", show.Rating, filters.MinRating),
		})
	}

	// Check minimum votes filter
	if filters.MinVotes > 0 {
		if show.Votes < filters.MinVotes {
			result.Checks = append(result.Checks, FilterCheck{
				Name:    "Minimum Votes",
				Passed:  false,
				Message: fmt.Sprintf("Votes %d below minimum %d", show.Votes, filters.MinVotes),
			})
			result.Passed = false
			result.Reason = fmt.Sprintf("votes %d below minimum %d", show.Votes, filters.MinVotes)
			return result
		}
		result.Checks = append(result.Checks, FilterCheck{
			Name:    "Minimum Votes",
			Passed:  true,
			Message: fmt.Sprintf("Votes %d meets minimum %d", show.Votes, filters.MinVotes),
		})
	}

	return result
}

// ShowPassesFilters checks if a show passes all configured filters
func ShowPassesFilters(show integrations.Show, filters config.ShowFilters) (bool, string) {
	result := ShowPassesFiltersDetailed(show, filters)
	return result.Passed, result.Reason
}

// Helper functions

func containsIgnoreCase(slice []string, item string) bool {
	itemLower := strings.ToLower(item)
	for _, s := range slice {
		if strings.ToLower(s) == itemLower {
			return true
		}
	}
	return false
}

func firstMatch(wanted, actual []string) string {
	for _, value := range actual {
		if containsIgnoreCase(wanted, value) {
			return value
		}
	}
	return ""
}

func matchingKeyword(keywords []string, title string) string {
	title = strings.ToLower(title)
	for _, keyword := range keywords {
		if keyword != "" && strings.Contains(title, strings.ToLower(keyword)) {
			return keyword
		}
	}
	return ""
}

func requiredChecks(genres []string, title, network string, requiredGenres, requiredKeywords, requiredNetworks []string) ([]FilterCheck, *FilterResult) {
	checks := make([]FilterCheck, 0, 3)
	if len(requiredGenres) > 0 {
		match := firstMatch(requiredGenres, genres)
		if match == "" {
			return nil, failed("Required genre", "Required genre not matched: "+strings.Join(requiredGenres, ", "))
		}
		checks = append(checks, FilterCheck{Name: "Required genre", Passed: true, Message: "Required genre matched: " + match})
	}
	if len(requiredKeywords) > 0 {
		match := matchingKeyword(requiredKeywords, title)
		if match == "" {
			return nil, failed("Required keyword", "Required title keyword not matched: "+strings.Join(requiredKeywords, ", "))
		}
		checks = append(checks, FilterCheck{Name: "Required keyword", Passed: true, Message: "Required title keyword matched: " + match})
	}
	if len(requiredNetworks) > 0 {
		if !containsIgnoreCase(requiredNetworks, network) {
			return nil, failed("Required network", "Required network not matched: "+strings.Join(requiredNetworks, ", "))
		}
		checks = append(checks, FilterCheck{Name: "Required network", Passed: true, Message: "Required network matched: " + network})
	}
	if len(checks) == 0 {
		return nil, nil
	}
	return checks, nil
}
