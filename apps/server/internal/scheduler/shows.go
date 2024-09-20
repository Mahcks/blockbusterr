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

// AnticipatedShowJobFunc handles fetching and processing anticipated shows
func (s Scheduler) AnticipatedShowJobFunc() {
	err := s.helpers.Trakt.Ping(context.Background())
	if err != nil {
		log.Warn("[scheduler] Skipping antcipated show job because of missing Trakt client ID.")
		return
	}

	log.Info("[scheduler] Running anticipated shows job...")

	startTime := time.Now()
	sj := sonarrJob{}
	gctx := s.gctx

	// Get Ombi enabled setting
	currentMode, err := gctx.Crate().SQL.Queries().GetSettingByKey(gctx, structures.SettingMode.String())
	if err != nil {
		log.Error("[show-job] Error getting Ombi enabled setting", "error", err)
		return
	}

	var ombiEnabled string
	if currentMode.Value.String == "ombi" {
		ombiEnabled = "true"
	} else {
		ombiEnabled = "false"
	}

	// Get Sonarr and Show settings
	sj.sonarrSettings, sj.showSettings, err = getSonarrAndShowSettings(gctx)
	if err != nil {
		return
	}

	// Fetch Anticipated Shows
	if sj.showSettings.Anticipated.Valid && sj.showSettings.Anticipated.Int32 > 0 {
		params := buildTraktParamsFromShowSettings(sj.showSettings, 1000, true)
		anticipatedShows, err := s.helpers.Trakt.GetAnticipatedShows(gctx, params)
		if err != nil {
			if errors.Is(err, trakt.ErrNoTraktSettings) {
				log.Warn("[show-job] Couldn't complete the job because Trakt client ID isn't set!")
			} else {
				log.Error("[show-job] Error fetching anticipated shows from Trakt", "error", err)
			}
		} else {
			sj.anticipatedShows = filterAndLimitShows(extractShowsFromAnticipated(anticipatedShows), sj.showSettings, int(sj.showSettings.Anticipated.Int32))
		}
	}

	// Process Ombi or Sonarr
	processShows(s, s.helpers, sj.anticipatedShows, sj.sonarrSettings, sj.ombiSettings, ombiEnabled, "Anticipated")

	log.Infof("[scheduler] Completed anticipated shows job in %.2f seconds!", time.Since(startTime).Seconds())
}

// PopularShowJobFunc handles fetching and processing popular shows
func (s Scheduler) PopularShowJobFunc() {
	err := s.helpers.Trakt.Ping(context.Background())
	if err != nil {
		log.Warn("[scheduler] Skipping popular show job because of missing Trakt client ID.")
		return
	}

	log.Info("[scheduler] Running popular shows job...")

	startTime := time.Now()
	sj := sonarrJob{}

	// Get Ombi enabled setting
	currentMode, err := s.gctx.Crate().SQL.Queries().GetSettingByKey(s.gctx, structures.SettingMode.String())
	if err != nil {
		log.Error("[show-job] Error getting Ombi enabled setting", "error", err)
		return
	}

	var ombiEnabled string
	if currentMode.Value.String == "ombi" {
		ombiEnabled = "true"
	} else {
		ombiEnabled = "false"
	}

	// Get Sonarr and Show settings
	sj.sonarrSettings, sj.showSettings, err = getSonarrAndShowSettings(s.gctx)
	if err != nil {
		return
	}

	// Fetch Popular Shows
	if sj.showSettings.Popular.Valid && sj.showSettings.Popular.Int32 > 0 {
		params := buildTraktParamsFromShowSettings(sj.showSettings, 1000, false)
		popularShows, err := s.helpers.Trakt.GetPopularShows(s.gctx, params)
		if err != nil {
			if errors.Is(err, trakt.ErrNoTraktSettings) {
				log.Warn("[show-job] Couldn't complete the job because Trakt client ID isn't set!")
			} else {
				log.Error("[show-job] Error fetching popular shows from Trakt", "error", err)
			}
		} else {
			sj.popularShows = filterAndLimitShows(extractShowsFromPopular(popularShows), sj.showSettings, int(sj.showSettings.Popular.Int32))
		}
	}

	// Process Ombi or Sonarr
	processShows(s, s.helpers, sj.popularShows, sj.sonarrSettings, sj.ombiSettings, ombiEnabled, "Popular")

	log.Infof("[scheduler] Completed popular shows job in %.2f seconds!", time.Since(startTime).Seconds())
}

// TrendingShowJobFunc handles fetching and processing trending shows
func (s Scheduler) TrendingShowJobFunc() {
	err := s.helpers.Trakt.Ping(context.Background())
	if err != nil {
		log.Warn("[scheduler] Skipping trending show job because of missing Trakt client ID.")
		return
	}

	log.Info("[scheduler] Running trending shows job...")

	startTime := time.Now()
	sj := sonarrJob{}

	// Get Ombi enabled setting
	// Get Ombi enabled setting
	currentMode, err := s.gctx.Crate().SQL.Queries().GetSettingByKey(s.gctx, structures.SettingMode.String())
	if err != nil {
		log.Error("[show-job] Error getting Ombi enabled setting", "error", err)
		return
	}

	var ombiEnabled string
	if currentMode.Value.String == "ombi" {
		ombiEnabled = "true"
	} else {
		ombiEnabled = "false"
	}

	// Get Sonarr and Show settings
	sj.sonarrSettings, sj.showSettings, err = getSonarrAndShowSettings(s.gctx)
	if err != nil {
		if errors.Is(err, db.ErrNoShowSettings) {
			log.Warn("[show-job] Skipping Sonarr job because of missing Show settings.")
		}
		return
	}

	// Fetch Trending Shows
	if sj.showSettings.Trending.Valid && sj.showSettings.Trending.Int32 > 0 {
		params := buildTraktParamsFromShowSettings(sj.showSettings, 1000, false)
		trendingShows, err := s.helpers.Trakt.GetTrendingShows(s.gctx, params)
		if err != nil {
			if errors.Is(err, trakt.ErrNoTraktSettings) {
				log.Warn("[show-job] Couldn't complete the job because Trakt client ID isn't set!")
			} else {
				log.Error("[show-job] Error fetching trending shows from Trakt", "error", err)
			}
		} else {
			sj.trendingShows = filterAndLimitShows(extractShowsFromTrending(trendingShows), sj.showSettings, int(sj.showSettings.Trending.Int32))
		}
	}

	// Process Ombi or Sonarr
	processShows(s, s.helpers, sj.trendingShows, sj.sonarrSettings, sj.ombiSettings, ombiEnabled, "Trending")

	log.Infof("[scheduler] Completed trending shows job in %.2f seconds!", time.Since(startTime).Seconds())
}

// Helper function to get Sonarr and Show settings
func getSonarrAndShowSettings(gctx global.Context) (db.SonarrSettings, db.ShowSettings, error) {
	sj := sonarrJob{}
	var err error

	// Get all settings from Sonarr table
	sj.sonarrSettings, err = gctx.Crate().SQL.Queries().GetSonarrSettings(gctx)
	if err != nil {
		if errors.Is(err, db.ErrNoSonarrSettings) {
			log.Warn("[sonarr-job] Skipping Sonarr job because of missing Sonarr settings.")
			return sj.sonarrSettings, sj.showSettings, nil
		} else {
			log.Errorf("[sonarr-job] Error getting Sonarr settings: %v", err)

			return sj.sonarrSettings, sj.showSettings, err
		}
	}

	// Get all the settings for shows
	sj.showSettings, err = gctx.Crate().SQL.Queries().GetShowSettings(gctx)
	if err != nil {
		if errors.Is(err, db.ErrNoShowSettings) {
			log.Warn("[sonarr-job] Skipping Sonarr job because of missing Show settings.")
			return sj.sonarrSettings, sj.showSettings, nil
		} else {
			log.Error("[sonarr-job] Error getting show settings", "error", err)
			return sj.sonarrSettings, sj.showSettings, err
		}
	}

	return sj.sonarrSettings, sj.showSettings, nil
}

// Helper function to process shows (Ombi or Sonarr)
func processShows(s Scheduler, helpers helpers.Helpers, shows []trakt.Show, sonarrSettings db.SonarrSettings, ombiSettings db.OmbiSettings, ombiEnabled string, jobType string) {
	if ombiEnabled == "true" {
		// If Ombi is enabled, request shows via Ombi
		requestShowsToOmbi(helpers.Ombi, s.notifications, shows, ombiSettings)
	} else {
		// Otherwise, request shows via Sonarr
		requestShowsToSonarr(s.gctx, helpers, s.notifications, shows, sonarrSettings)
	}

	log.Infof("[scheduler] %s shows processed. Total: %d", jobType, len(shows))
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
	qualityProfiles, err := r.GetQualityProfiles(nil, nil)
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
	rootFolders, err := r.GetRootFolders(nil, nil)
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

func requestShowsToSonarr(gctx global.Context, helpers helpers.Helpers, notifications *notifications.NotificationManager, shows []trakt.Show, sonarrSettings db.SonarrSettings) {
	// Fetch quality profile and root folder from Sonarr
	qualityProfileID, rootFolderPath, err := fetchSonarrSettings(helpers.Sonarr, sonarrSettings)
	if err != nil {
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
		_, err := helpers.Sonarr.RequestSeries(context.Background(), nil, nil, body)
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

			// Get show poster
			media, err := helpers.OMDb.GetMedia(context.Background(), show.IDs.IMDB)
			if err != nil {
				log.Errorf("[sonarr-job] Failed to get show poster for %s: %v", show.Title, err)
				continue
			}

			// Add show to recently added
			err = gctx.Crate().SQL.Queries().AddToRecentlyAddedMedia(
				context.Background(),
				"SHOW",
				media.Title,
				show.Year,
				media.Plot,
				media.IMDBID,
				media.Poster,
			)
			if err != nil {
				log.Errorf("[sonarr-job] Failed to add show %s to recently added: %v", show.Title, err)
				continue
			}

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
