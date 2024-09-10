package scheduler

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/mahcks/blockbusterr/internal/db"
	"github.com/mahcks/blockbusterr/internal/global"
	"github.com/mahcks/blockbusterr/internal/helpers"
	"github.com/mahcks/blockbusterr/internal/helpers/sonarr"
	"github.com/mahcks/blockbusterr/internal/helpers/trakt"
	"github.com/mahcks/blockbusterr/pkg/utils"
)

type sonarrJob struct {
	sonarrSettings db.SonarrSettings
	showSettings   db.ShowSettings

	anticipatedShows []trakt.Show
	popularShows     []trakt.Show
	trendingShows    []trakt.Show
}

// SonarrJobFunc defines the logic for the Sonarr job
func (s Scheduler) SonarrJobFunc(gctx global.Context, helpers helpers.Helpers) {
	log.Info("[scheduler] Running show job...")

	// Start time tracking
	startTime := time.Now()

	sj := sonarrJob{}
	var err error

	// Step 1. Get all settings from Sonarr table
	sj.sonarrSettings, err = gctx.Crate().SQL.Queries().GetSonarrSettings(gctx)
	if err != nil {
		log.Errorf("[sonarr-job] Error getting Sonarr settings: %v", err)
		return
	}

	// Step 2. Get all the settings for shows
	sj.showSettings, err = gctx.Crate().SQL.Queries().GetShowSettings(gctx)
	if err != nil {
		log.Error("[sonarr-job] Error getting show settings", "error", err)
		return
	}

	// Query a large number of shows from each list (e.g., 1000)
	largeShowQueryLimit := 1000

	utils.PrettyPrintStruct(sj.showSettings)

	// Fetch Anticipated Shows
	if sj.showSettings.Anticipated.Valid {
		if sj.showSettings.Anticipated.Int32 == 0 {
			log.Warn("[sonarr-job] Anticipated shows are enabled but the limit is set to 0. Skipping...")
		} else {
			params := buildTraktParamsFromShowSettings(sj.showSettings, largeShowQueryLimit)
			anticipatedShows, err := helpers.Trakt.GetAnticipatedShows(gctx, params)

			if err != nil {
				log.Error("[sonarr-job] Error fetching anticipated shows from Trakt", "error", err)
			} else {
				shows := extractShowsFromAnticipated(anticipatedShows)
				filteredShows := applyAdditionalFiltersToShows(shows, sj.showSettings)
				sj.anticipatedShows = getTopNShows(filteredShows, int(sj.showSettings.Anticipated.Int32))

				// Make requests to Sonarr for Anticipated shows
				requestShowsToSonarr(helpers.Sonarr, sj.anticipatedShows, sj.sonarrSettings)
			}
		}
	}

	// Fetch Popular Shows
	if sj.showSettings.Popular.Valid {
		if sj.showSettings.Popular.Int32 == 0 {
			log.Warn("[sonarr-job] Popular shows are enabled but the limit is set to 0. Skipping...")
		} else {
			params := buildTraktParamsFromShowSettings(sj.showSettings, largeShowQueryLimit)
			popularShows, err := helpers.Trakt.GetPopularShows(gctx, params)
			if err != nil {
				log.Error("[sonarr-job] Error fetching popular shows from Trakt", "error", err)
			} else {
				shows := extractShowsFromPopular(popularShows)
				filteredShows := applyAdditionalFiltersToShows(shows, sj.showSettings)
				sj.popularShows = getTopNShows(filteredShows, int(sj.showSettings.Popular.Int32))

				// Make requests to Sonarr for popular shows
				requestShowsToSonarr(helpers.Sonarr, sj.popularShows, sj.sonarrSettings)
			}
		}
	}

	// Fetch Trending Shows
	if sj.showSettings.Trending.Valid {
		if sj.showSettings.Trending.Int32 == 0 {
			log.Warn("[sonarr-job] Trending shows are enabled but the limit is set to 0. Skipping...")
		} else {
			params := buildTraktParamsFromShowSettings(sj.showSettings, largeShowQueryLimit)
			trendingShows, err := helpers.Trakt.GetTrendingShows(gctx, params)
			if err != nil {
				log.Error("[sonarr-job] Error fetching trending shows from Trakt", "error", err)
			} else {
				shows := extractShowsFromTrending(trendingShows)
				filteredShows := applyAdditionalFiltersToShows(shows, sj.showSettings)
				sj.trendingShows = getTopNShows(filteredShows, int(sj.showSettings.Trending.Int32))

				// Make requests to Sonarr for trending shows
				requestShowsToSonarr(helpers.Sonarr, sj.trendingShows, sj.sonarrSettings)
			}
		}
	}

	log.Debug("[sonarr-job] Anticipated Shows", "count", len(sj.anticipatedShows))
	log.Debug("[sonarr-job] Popular Shows", "count", len(sj.popularShows))
	log.Debug("[sonarr-job] Trending Shows", "count", len(sj.trendingShows))

	duration := time.Since(startTime)
	durationInSeconds := float64(duration.Milliseconds()) / 1000
	roundedDuration := fmt.Sprintf("%.2f", durationInSeconds)

	log.Infof("[scheduler] Completed Sonarr job in %v seconds!", roundedDuration)
}

// Helper function to build Trakt API request parameters from the show settings
func buildTraktParamsFromShowSettings(settings db.ShowSettings, limit int) *trakt.TraktMovieParams {
	params := &trakt.TraktMovieParams{}
	params.Extended = "full"

	// Set the limit based on the settings
	params.Limit = limit

	// Build allowed countries string (comma-separated)
	allowedCountries := []string{}
	for _, country := range settings.AllowedCountries {
		allowedCountries = append(allowedCountries, country.CountryCode)
	}
	if len(allowedCountries) > 0 {
		params.Countries = strings.Join(allowedCountries, ",")
	}

	// Build allowed languages string (comma-separated)
	allowedLanguages := []string{}
	for _, language := range settings.AllowedLanguages {
		allowedLanguages = append(allowedLanguages, language.LanguageCode)
	}
	if len(allowedLanguages) > 0 {
		params.Languages = strings.Join(allowedLanguages, ",")
	}

	// Apply runtime filter
	if settings.MinRuntime.Valid || settings.MaxRuntime.Valid {
		minRuntime := 0
		maxRuntime := 0

		if settings.MinRuntime.Valid {
			minRuntime = int(settings.MinRuntime.Int32)
		}
		if settings.MaxRuntime.Valid {
			maxRuntime = int(settings.MaxRuntime.Int32)
		}

		params.Runtime = fmt.Sprintf("%d-%d", minRuntime, maxRuntime)
	}

	// Apply year filter (if applicable)
	if settings.MinYear.Valid || settings.MaxYear.Valid {
		minYear := 0
		maxYear := 0

		if settings.MinYear.Valid {
			minYear = int(settings.MinYear.Int32)
		}
		if settings.MaxYear.Valid {
			maxYear = int(settings.MaxYear.Int32)
		} else {
			maxYear = time.Now().Year() // Default max year is the current year
		}

		params.Years = fmt.Sprintf("%d-%d", minYear, maxYear)
	}

	// Return the configured parameters
	return params
}

func applyAdditionalFiltersToShows(shows []trakt.Show, settings db.ShowSettings) []trakt.Show {
	filteredShows := []trakt.Show{}

	// Build blacklisted genres, keywords, and TVDB IDs from settings
	blacklistedGenres := map[string]bool{}
	for _, genre := range settings.BlacklistedGenres {
		blacklistedGenres[strings.ToLower(genre.Genre)] = true
	}

	blacklistedKeywords := map[string]bool{}
	for _, keyword := range settings.BlacklistedTitleKeywords {
		blacklistedKeywords[strings.ToLower(keyword.Keyword)] = true
	}

	blacklistedTVDBIDs := map[int]bool{}
	for _, tvdbID := range settings.BlacklistedTVDBIDs {
		blacklistedTVDBIDs[tvdbID.TVDBID] = true
	}

	// Loop through the shows and filter them
	for _, show := range shows {
		// Check if the show is blacklisted by TVDB ID
		if blacklistedTVDBIDs[show.IDs.TVDB] {
			log.Infof("Skipping show '%s' due to blacklisted TVDB ID: %d", show.Title, show.IDs.TVDB)
			continue // Skip this show and move to the next
		}

		// Check if the show is blacklisted by genre
		blacklisted := false
		for _, genre := range show.Genres {
			if blacklistedGenres[strings.ToLower(genre)] {
				blacklisted = true
				break
			}
		}
		if blacklisted {
			continue // Skip this show and move to the next
		}

		// Check if the show is blacklisted by title keywords
		for keyword := range blacklistedKeywords {
			if strings.Contains(strings.ToLower(show.Title), keyword) {
				log.Infof("Skipping show '%s' due to blacklisted keyword: %s", show.Title, keyword)
				blacklisted = true
				break
			}
		}
		if blacklisted {
			continue // Skip this show and move to the next
		}

		// If the show passed all filters, add it to the filtered list
		filteredShows = append(filteredShows, show)
	}

	return filteredShows
}

func getTopNShows(shows []trakt.Show, limit int) []trakt.Show {
	if len(shows) <= limit {
		return shows
	}
	return shows[:limit]
}

// Extract Shows from TrendingShows
func extractShowsFromTrending(trendingShows []trakt.TrendingShow) []trakt.Show {
	shows := []trakt.Show{}
	for _, trendingShow := range trendingShows {
		shows = append(shows, trendingShow.Show)
	}
	return shows
}

// Extract Shows from PopularShows
func extractShowsFromPopular(popularShows []trakt.Show) []trakt.Show {
	return popularShows
}

// Extract Shows from AnticipatedShows
func extractShowsFromAnticipated(anticipatedShows []trakt.AnticipatedShow) []trakt.Show {
	shows := []trakt.Show{}
	for _, anticipated := range anticipatedShows {
		shows = append(shows, anticipated.Show)
	}
	return shows
}

func requestShowsToSonarr(s sonarr.Service, shows []trakt.Show, sonarrSettings db.SonarrSettings) {
	// Fetch quality profile and root folder from Sonarr
	qualityProfileID, rootFolderPath, err := fetchSonarrSettings(s, sonarrSettings)
	if err != nil {
		log.Error("[sonarr-job: sonarr] Error fetching Sonarr settings", "error", err)
		return
	}

	for _, show := range shows {
		// Prepare the request body for Sonarr

		body := sonarr.RequestSeriesBody{
			Title:            show.Title,
			TVDbId:           show.IDs.TVDB,
			Monitored:        true,
			QualityProfileID: qualityProfileID,
			RootFolderPath:   rootFolderPath,
		}

		body.AddOptions.SearchForMissingEpisodes = true

		utils.PrettyPrintStruct(body)

		// Make the request to Sonarr
		/* _, err := s.RequestSeries(context.Background(), body)
		if err != nil {
			if errors.Is(err, sonarr.ErrShowAlreadyExists) {
				// Log a warning if the show already exists in Sonarr
				log.Warnf(`[sonarr-job] Skipping "%s" as it already exists in Sonarr...`, show.Title)
			} else {
				// Log an error for any other issues
				log.Errorf("[sonarr-job] Failed to request show %s: %v", show.Title, err)
			}
		} else {
			// Log a success message if the show was added successfully
			log.Infof("[sonarr-job] Show requested successfully: %s", show.Title)
		} */
	}
}

func fetchSonarrSettings(r sonarr.Service, sonarrSettings db.SonarrSettings) (int, string, error) {
	// Get quality profiles from Sonarr
	qualityProfiles, err := r.GetQualityProfiles()
	if err != nil {
		return 0, "", fmt.Errorf("failed to get Sonarr quality profiles: %w", err)
	}

	// Find the quality profile matching the stored quality ID
	var qualityProfileID int
	for _, profile := range qualityProfiles {
		if profile.ID == int(sonarrSettings.Quality.Int32) {
			qualityProfileID = profile.ID
			break
		}
	}
	if qualityProfileID == 0 {
		return 0, "", fmt.Errorf("quality profile ID %v not found", sonarrSettings.Quality.Int32)
	}

	// Get root folders from Sonarr
	rootFolders, err := r.GetRootFolders()
	if err != nil {
		return 0, "", fmt.Errorf("failed to get Sonarr root folders: %w", err)
	}

	// Find the root folder matching the stored root folder ID
	var rootFolderPath string
	for _, folder := range rootFolders {
		if folder.ID == int(sonarrSettings.RootFolder.Int32) {
			rootFolderPath = folder.Path
			break
		}
	}
	if rootFolderPath == "" {
		return 0, "", fmt.Errorf("root folder ID %v not found", sonarrSettings.RootFolder.Int32)
	}

	// Return the matched quality profile ID and root folder path
	return qualityProfileID, rootFolderPath, nil
}
