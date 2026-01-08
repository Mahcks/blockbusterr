package filters

import (
	"strings"

	"github.com/gofiber/fiber/v2/log"
	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/integrations"
)

// MoviePassesFilters checks if a movie passes all configured filters
func MoviePassesFilters(movie integrations.Movie, filters config.MovieFilters) (bool, string) {
	// Check TMDB ID blacklist
	for _, id := range filters.BlacklistedTMDBIds {
		if movie.IDs.TMDB == id {
			return false, "blacklisted TMDB ID"
		}
	}

	// Check country filter
	if len(filters.AllowedCountries) > 0 && !contains(filters.AllowedCountries, "ignore") {
		// Only filter if country is specified - allow through if missing (Trakt doesn't always provide this)
		if movie.Country != "" && !containsIgnoreCase(filters.AllowedCountries, movie.Country) {
			return false, "country not allowed: " + movie.Country
		}
	}

	// Check language filter
	if len(filters.AllowedLanguages) > 0 && !contains(filters.AllowedLanguages, "ignore") {
		// Only filter if language is specified - allow through if missing (Trakt doesn't always provide this)
		if movie.Language != "" && !containsIgnoreCase(filters.AllowedLanguages, movie.Language) {
			return false, "language not allowed: " + movie.Language
		}
	}

	// Check genre blacklist
	if len(filters.BlacklistedGenres) > 0 && !contains(filters.BlacklistedGenres, "ignore") {
		for _, genre := range movie.Genres {
			if containsIgnoreCase(filters.BlacklistedGenres, genre) {
				return false, "blacklisted genre: " + genre
			}
		}
	}

	// Check title keywords blacklist
	if len(filters.BlacklistedKeywords) > 0 {
		titleLower := strings.ToLower(movie.Title)
		for _, keyword := range filters.BlacklistedKeywords {
			if strings.Contains(titleLower, strings.ToLower(keyword)) {
				return false, "blacklisted keyword in title: " + keyword
			}
		}
	}

	// Check runtime filters
	if movie.Runtime > 0 {
		if filters.BlacklistedMinRuntime > 0 && movie.Runtime < filters.BlacklistedMinRuntime {
			return false, "runtime too short"
		}
		if filters.BlacklistedMaxRuntime > 0 && movie.Runtime > filters.BlacklistedMaxRuntime {
			return false, "runtime too long"
		}
	}

	// Check year filters
	if movie.Year > 0 {
		if filters.BlacklistedMinYear > 0 && movie.Year < filters.BlacklistedMinYear {
			return false, "year too old"
		}
		if filters.BlacklistedMaxYear > 0 && movie.Year > filters.BlacklistedMaxYear {
			return false, "year too new"
		}
	}

	return true, ""
}

// ShowPassesFilters checks if a show passes all configured filters
func ShowPassesFilters(show integrations.Show, filters config.ShowFilters) (bool, string) {
	// Check TVDB ID blacklist
	for _, id := range filters.BlacklistedTVDBIds {
		if show.IDs.TVDB == id {
			return false, "blacklisted TVDB ID"
		}
	}

	// Check country filter
	if len(filters.AllowedCountries) > 0 && !contains(filters.AllowedCountries, "ignore") {
		// Only filter if country is specified - allow through if missing (Trakt doesn't always provide this)
		if show.Country != "" && !containsIgnoreCase(filters.AllowedCountries, show.Country) {
			return false, "country not allowed: " + show.Country
		}
	}

	// Check language filter
	if len(filters.AllowedLanguages) > 0 && !contains(filters.AllowedLanguages, "ignore") {
		// Only filter if language is specified - allow through if missing (Trakt doesn't always provide this)
		if show.Language != "" && !containsIgnoreCase(filters.AllowedLanguages, show.Language) {
			return false, "language not allowed: " + show.Language
		}
	}

	// Check genre blacklist
	if len(filters.BlacklistedGenres) > 0 && !contains(filters.BlacklistedGenres, "ignore") {
		for _, genre := range show.Genres {
			if containsIgnoreCase(filters.BlacklistedGenres, genre) {
				return false, "blacklisted genre: " + genre
			}
		}
	}

	// Check network blacklist
	if len(filters.BlacklistedNetworks) > 0 {
		if show.Network != "" && containsIgnoreCase(filters.BlacklistedNetworks, show.Network) {
			return false, "blacklisted network: " + show.Network
		}
	}

	// Check title keywords blacklist
	if len(filters.BlacklistedKeywords) > 0 {
		titleLower := strings.ToLower(show.Title)
		for _, keyword := range filters.BlacklistedKeywords {
			if strings.Contains(titleLower, strings.ToLower(keyword)) {
				return false, "blacklisted keyword in title: " + keyword
			}
		}
	}

	// Check runtime filters
	if show.Runtime > 0 {
		if filters.BlacklistedMinRuntime > 0 && show.Runtime < filters.BlacklistedMinRuntime {
			return false, "runtime too short"
		}
		if filters.BlacklistedMaxRuntime > 0 && show.Runtime > filters.BlacklistedMaxRuntime {
			return false, "runtime too long"
		}
	}

	// Check year filters
	if show.Year > 0 {
		if filters.BlacklistedMinYear > 0 && show.Year < filters.BlacklistedMinYear {
			return false, "year too old"
		}
		if filters.BlacklistedMaxYear > 0 && show.Year > filters.BlacklistedMaxYear {
			return false, "year too new"
		}
	}

	return true, ""
}

// FilterMovies filters a list of movies based on configured filters
func FilterMovies(movies []integrations.Movie, filters config.MovieFilters) []integrations.Movie {
	if !hasAnyFilters(filters) {
		return movies
	}

	filtered := make([]integrations.Movie, 0)
	for _, movie := range movies {
		passes, reason := MoviePassesFilters(movie, filters)
		if passes {
			filtered = append(filtered, movie)
		} else {
			log.Debugf("Filtered out movie '%s' (%d): %s", movie.Title, movie.Year, reason)
		}
	}
	return filtered
}

// FilterShows filters a list of shows based on configured filters
func FilterShows(shows []integrations.Show, filters config.ShowFilters) []integrations.Show {
	if !hasAnyShowFilters(filters) {
		return shows
	}

	filtered := make([]integrations.Show, 0)
	for _, show := range shows {
		passes, reason := ShowPassesFilters(show, filters)
		if passes {
			filtered = append(filtered, show)
		} else {
			log.Debugf("Filtered out show '%s' (%d): %s", show.Title, show.Year, reason)
		}
	}
	return filtered
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func containsIgnoreCase(slice []string, item string) bool {
	itemLower := strings.ToLower(item)
	for _, s := range slice {
		if strings.ToLower(s) == itemLower {
			return true
		}
	}
	return false
}

func hasAnyFilters(filters config.MovieFilters) bool {
	return len(filters.AllowedCountries) > 0 ||
		len(filters.AllowedLanguages) > 0 ||
		len(filters.BlacklistedGenres) > 0 ||
		len(filters.BlacklistedKeywords) > 0 ||
		len(filters.BlacklistedTMDBIds) > 0 ||
		filters.BlacklistedMinRuntime > 0 ||
		filters.BlacklistedMaxRuntime > 0 ||
		filters.BlacklistedMinYear > 0 ||
		filters.BlacklistedMaxYear > 0
}

func hasAnyShowFilters(filters config.ShowFilters) bool {
	return len(filters.AllowedCountries) > 0 ||
		len(filters.AllowedLanguages) > 0 ||
		len(filters.BlacklistedGenres) > 0 ||
		len(filters.BlacklistedKeywords) > 0 ||
		len(filters.BlacklistedNetworks) > 0 ||
		len(filters.BlacklistedTVDBIds) > 0 ||
		filters.BlacklistedMinRuntime > 0 ||
		filters.BlacklistedMaxRuntime > 0 ||
		filters.BlacklistedMinYear > 0 ||
		filters.BlacklistedMaxYear > 0
}
