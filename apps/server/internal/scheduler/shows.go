package scheduler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/mahcks/blockbusterr/internal/db"
	"github.com/mahcks/blockbusterr/internal/global"
	"github.com/mahcks/blockbusterr/internal/helpers"
	"github.com/mahcks/blockbusterr/internal/helpers/ombi"
	"github.com/mahcks/blockbusterr/internal/helpers/sonarr"
	"github.com/mahcks/blockbusterr/internal/helpers/trakt"
	"github.com/mahcks/blockbusterr/internal/notifications"
	"github.com/mahcks/blockbusterr/pkg/structures"
)

type sonarrJob struct {
	ombiSettings   db.OmbiSettings
	sonarrSettings db.SonarrSettings
	showSettings   db.ShowSettings

	anticipatedShows []trakt.Show
	popularShows     []trakt.Show
	trendingShows    []trakt.Show
}

// ShowJobFunc defines the logic for the TV show job
func (s Scheduler) ShowJobFunc(gctx global.Context, helpers helpers.Helpers) {
	log.Info("[scheduler] Running show job...")

	// Start time tracking
	startTime := time.Now()

	sj := sonarrJob{}
	var err error

	ombiEnabled, err := gctx.Crate().SQL.Queries().GetSettingByKey(gctx, structures.SettingOmbiEnabled.String())
	if err != nil {
		log.Error("[show-job] Error getting Ombi enabled setting", "error", err)
		return
	}

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

	// Fetch Anticipated Shows
	if sj.showSettings.Anticipated.Valid && sj.showSettings.Anticipated.Int32 > 0 {
		params := buildTraktParamsFromShowSettings(sj.showSettings, largeShowQueryLimit, true)
		anticipatedShows, err := helpers.Trakt.GetAnticipatedShows(gctx, params)
		if err != nil {
			log.Error("[show-job] Error fetching anticipated shows from Trakt", "error", err)
		} else {
			sj.anticipatedShows = filterAndLimitShows(extractShowsFromAnticipated(anticipatedShows), sj.showSettings, int(sj.showSettings.Anticipated.Int32))
		}
	}

	// Fetch Popular Shows
	if sj.showSettings.Popular.Valid && sj.showSettings.Popular.Int32 > 0 {
		params := buildTraktParamsFromShowSettings(sj.showSettings, largeShowQueryLimit, false)
		popularShows, err := helpers.Trakt.GetPopularShows(gctx, params)
		if err != nil {
			log.Error("[show-job] Error fetching popular shows from Trakt", "error", err)
		} else {
			sj.popularShows = filterAndLimitShows(extractShowsFromPopular(popularShows), sj.showSettings, int(sj.showSettings.Popular.Int32))
		}
	}

	// Fetch Trending Shows
	if sj.showSettings.Trending.Valid && sj.showSettings.Trending.Int32 > 0 {
		params := buildTraktParamsFromShowSettings(sj.showSettings, largeShowQueryLimit, false)
		trendingShows, err := helpers.Trakt.GetTrendingShows(gctx, params)
		if err != nil {
			log.Error("[show-job] Error fetching trending shows from Trakt", "error", err)
		} else {
			sj.trendingShows = filterAndLimitShows(extractShowsFromTrending(trendingShows), sj.showSettings, int(sj.showSettings.Trending.Int32))
		}
	}

	// If Ombi is enabled, use Ombi settings and request shows via Ombi
	if ombiEnabled.Value.String == "true" {
		sj.ombiSettings, err = gctx.Crate().SQL.Queries().GetOmbiSettings(gctx)
		if err != nil {
			log.Error("[show-job] Error getting Ombi settings", "error", err)
			return
		}

		// Request shows via Ombi
		requestShowsToOmbi(helpers.Ombi, s.notifications, sj.anticipatedShows, sj.ombiSettings)
		requestShowsToOmbi(helpers.Ombi, s.notifications, sj.popularShows, sj.ombiSettings)
		requestShowsToOmbi(helpers.Ombi, s.notifications, sj.trendingShows, sj.ombiSettings)

		log.Debug("[ombi-job] Anticipated Shows", "count", len(sj.anticipatedShows))
		log.Debug("[ombi-job] Popular Shows", "count", len(sj.popularShows))
		log.Debug("[ombi-job] Trending Shows", "count", len(sj.trendingShows))

	} else {
		// If Ombi is not enabled, fallback to Sonarr
		requestShowsToSonarr(helpers.Sonarr, s.notifications, sj.anticipatedShows, sj.sonarrSettings)
		requestShowsToSonarr(helpers.Sonarr, s.notifications, sj.popularShows, sj.sonarrSettings)
		requestShowsToSonarr(helpers.Sonarr, s.notifications, sj.trendingShows, sj.sonarrSettings)

		log.Debug("[sonarr-job] Anticipated Shows", "count", len(sj.anticipatedShows))
		log.Debug("[sonarr-job] Popular Shows", "count", len(sj.popularShows))
		log.Debug("[sonarr-job] Trending Shows", "count", len(sj.trendingShows))
	}

	duration := time.Since(startTime)
	durationInSeconds := float64(duration.Milliseconds()) / 1000
	roundedDuration := fmt.Sprintf("%.2f", durationInSeconds)

	log.Infof("[scheduler] Completed Sonarr job in %v seconds!", roundedDuration)
}

// Helper function to build Trakt API request parameters from the show settings
func buildTraktParamsFromShowSettings(settings db.ShowSettings, limit int, isAnticipated bool) *trakt.TraktMovieParams {
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

	// Apply year filter (with custom logic for anticipated items)
	if isAnticipated {
		// For anticipated shows or movies, allow future years
		minYear := time.Now().Year()
		maxYear := minYear + 10 // Set a future year limit

		if settings.MinYear.Valid {
			minYear = int(settings.MinYear.Int32)
		}
		// Don't apply MaxYear for anticipated items
		params.Years = fmt.Sprintf("%d-%d", minYear, maxYear)
	} else {
		// Regular logic for non-anticipated shows/movies
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

func filterAndLimitShows(shows []trakt.Show, settings db.ShowSettings, limit int) []trakt.Show {
	// Apply additional filters like allowed countries, allowed languages, and blacklists
	filteredShows := applyAdditionalFiltersToShows(shows, settings)

	// Limit the number of shows to the specified limit
	return getTopNShows(filteredShows, limit)
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

func requestShowsToOmbi(o ombi.Service, notifications *notifications.NotificationManager, shows []trakt.Show, ombiSettings db.OmbiSettings) {
	for _, show := range shows {
		body := ombi.RequestShowBody{
			TheMovieDBID: show.IDs.TMDB,
			RequestAll:   true,
			LanguageCode: "en",
		}

		// Set the request on behalf of a specific user if configured
		if ombiSettings.UserID.Valid && ombiSettings.UserID.String != "" {
			body.RequestOnBehalf = ombiSettings.UserID.String
		}

		// Set the quality profile override if configured
		if ombiSettings.ShowRootFolder.Valid || ombiSettings.ShowRootFolder.Int32 != 0 {
			rootFolder := fmt.Sprintf("%d", ombiSettings.ShowRootFolder.Int32)
			body.RootFolderOverride = &rootFolder
		}

		// Set the quality profile override if configured
		if ombiSettings.ShowQuality.Valid || ombiSettings.ShowQuality.Int32 != 0 {
			qualityProfile := fmt.Sprintf("%d", ombiSettings.ShowQuality.Int32)
			body.QualityPathOverride = &qualityProfile
		}

		// Request the show via Ombi
		_, err := o.RequestShow(body)
		if err != nil {
			if errors.Is(err, ombi.ErrShowAlreadyRequested) {
				log.Warnf(`[ombi-job] Skipping "%s" as it was already requested...`, show.Title)
			} else {
				log.Errorf("[ombi-job] Failed to request show %s via Ombi: %v", show.Title, err)
			}
		} else {
			log.Infof("[ombi-job] Show requested successfully via Ombi: %s", show.Title)
			showPayload, err := json.Marshal(show)
			if err != nil {
				log.Errorf("[ombi-job] Failed to marshal show payload: %v", err)
				continue
			}

			err = notifications.SendNotification(structures.SHOWADDEDALERT, showPayload)
			if err != nil {
				log.Errorf("[ombi-job] Failed to send notification for show %s: %v", show.Title, err)
			}
		}
	}
}

func requestShowsToSonarr(s sonarr.Service, notifications *notifications.NotificationManager, shows []trakt.Show, sonarrSettings db.SonarrSettings) {
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

		// Make the request to Sonarr
		_, err := s.RequestSeries(context.Background(), body)
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
			showPayload, err := json.Marshal(show)
			if err != nil {
				log.Errorf("[sonarr-job] Failed to marshal show payload: %v", err)
				continue
			}

			err = notifications.SendNotification(structures.SHOWADDEDALERT, showPayload)
			if err != nil {
				log.Errorf("[sonarr-job] Failed to send notification for show %s: %v", show.Title, err)
			}
		}
	}
}
