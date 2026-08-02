package jobs

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/database"
	"github.com/mahcks/blockbusterr/internal/filters"
	"github.com/mahcks/blockbusterr/internal/integrations"
)

// PreviewDynamicJob uses the same provider selection and fetchers as job execution.
func PreviewDynamicJob(cfg *config.Config, db *database.Database, job config.DynamicJob) (PreviewResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	definition, ok := GetJobTypeDefinition(job.Type)
	if !ok {
		return PreviewResponse{}, fmt.Errorf("unknown job type: %s", job.Type)
	}
	if job.Limit <= 0 {
		job.Limit = definition.DefaultLimit
	}
	cfg, err := configForJob(cfg, job)
	if err != nil {
		return PreviewResponse{}, err
	}
	mode := DetermineMode(job.Mode, cfg.Jobs.Mode)
	response := PreviewResponse{JobName: job.Name, Source: job.Source, MediaType: job.MediaType, Mode: mode, HasPosters: hasTMDBConfigured(cfg), Items: []PreviewItem{}}
	executor := &DynamicJobExecutor{Config: cfg, Database: db, DryRun: true}
	discovery, err := NewDiscoveryClient(cfg, job.Source)
	if err != nil {
		return response, err
	}

	if job.MediaType == "movie" {
		fetcher := executor.getMovieFetcher(job.Type)
		if fetcher == nil {
			return response, fmt.Errorf("no movie fetcher for job type: %s", job.Type)
		}
		movies, err := fetcher(ctx, discovery, job.Limit, job.Period)
		if err != nil {
			return response, err
		}
		response.TotalFound = len(movies)
		if err := previewMovies(ctx, cfg, mode, movies, &response); err != nil {
			return response, err
		}
		return response, nil
	}

	fetcher := executor.getShowFetcher(job.Type)
	if fetcher == nil {
		return response, fmt.Errorf("no show fetcher for job type: %s", job.Type)
	}
	shows, err := fetcher(ctx, discovery, job.Limit, job.Period)
	if err != nil {
		return response, err
	}
	response.TotalFound = len(shows)
	if err := previewShows(ctx, cfg, mode, shows, &response); err != nil {
		return response, err
	}
	return response, nil
}

func previewMovies(ctx context.Context, cfg *config.Config, mode string, movies []integrations.Movie, response *PreviewResponse) error {
	existing := map[int]bool{}
	if mode == "direct" {
		client := integrations.NewRadarr(integrations.RadarrConfig{BaseURL: cfg.Radarr.URL, APIKey: cfg.Radarr.APIKey})
		items, err := client.GetMovies(ctx)
		if err != nil {
			return fmt.Errorf("failed to check Radarr library: %w", err)
		}
		for _, item := range items {
			existing[item.TmdbID] = true
		}
	}
	var jellyseerr *integrations.Jellyseerr
	if mode == "jellyseerr" {
		jellyseerr = previewJellyseerr(cfg)
	}
	for index, movie := range movies {
		item := createMoviePreviewItem(cfg, movie, len(movies)-index)
		result := filters.MoviePassesRules(movie, cfg.Filters.Movies, cfg.TitleExceptions)
		item.FilterChecks = result.Checks
		item.DecisionReason = filters.Explain(result)
		if !result.Passed {
			item.FilteredOut, item.FilterReason = true, item.DecisionReason
			response.FilteredOut++
		} else if mode == "direct" && existing[movie.IDs.TMDB] {
			item.AlreadyExists = true
			item.DecisionReason = "Skipped: Already in Radarr"
			response.AlreadyExists++
		} else if jellyseerr != nil && movie.IDs.TMDB > 0 {
			info, err := jellyseerr.GetMovieInfo(movie.IDs.TMDB)
			if err != nil {
				return fmt.Errorf("failed to check Jellyseerr movie %d: %w", movie.IDs.TMDB, err)
			}
			if info.HasMediaInfo() {
				item.AlreadyExists = true
				item.DecisionReason = "Skipped: Already in Jellyseerr / Seerr"
				response.AlreadyExists++
			} else {
				response.WillAdd++
			}
		} else {
			item.DecisionReason = "Will add: " + strings.TrimPrefix(filters.Explain(result), "Accepted: ")
			response.WillAdd++
		}
		response.Items = append(response.Items, item)
	}
	return nil
}

func previewShows(ctx context.Context, cfg *config.Config, mode string, shows []integrations.Show, response *PreviewResponse) error {
	existingTVDB, existingTMDB := map[int]bool{}, map[int]bool{}
	if mode == "direct" {
		client := integrations.NewSonarr(integrations.SonarrConfig{BaseURL: cfg.Sonarr.URL, APIKey: cfg.Sonarr.APIKey})
		items, err := client.GetSeries(ctx)
		if err != nil {
			return fmt.Errorf("failed to check Sonarr library: %w", err)
		}
		for _, item := range items {
			existingTVDB[item.TvdbID] = true
			existingTMDB[item.TmdbID] = true
		}
	}
	var jellyseerr *integrations.Jellyseerr
	if mode == "jellyseerr" {
		jellyseerr = previewJellyseerr(cfg)
	}
	for index, show := range shows {
		item := createShowPreviewItem(cfg, show, len(shows)-index)
		result := filters.ShowPassesRules(show, cfg.Filters.Shows, cfg.TitleExceptions)
		item.FilterChecks = result.Checks
		item.DecisionReason = filters.Explain(result)
		if !result.Passed {
			item.FilteredOut, item.FilterReason = true, item.DecisionReason
			response.FilteredOut++
		} else if mode == "direct" && ((show.IDs.TVDB > 0 && existingTVDB[show.IDs.TVDB]) || (show.IDs.TMDB > 0 && existingTMDB[show.IDs.TMDB])) {
			item.AlreadyExists = true
			item.DecisionReason = "Skipped: Already in Sonarr"
			response.AlreadyExists++
		} else if jellyseerr != nil && show.IDs.TMDB > 0 {
			info, err := jellyseerr.GetShowInfo(show.IDs.TMDB)
			if err != nil {
				return fmt.Errorf("failed to check Jellyseerr show %d: %w", show.IDs.TMDB, err)
			}
			if info.HasMediaInfo() {
				item.AlreadyExists = true
				item.DecisionReason = "Skipped: Already in Jellyseerr / Seerr"
				response.AlreadyExists++
			} else {
				response.WillAdd++
			}
		} else {
			item.DecisionReason = "Will add: " + strings.TrimPrefix(filters.Explain(result), "Accepted: ")
			response.WillAdd++
		}
		response.Items = append(response.Items, item)
	}
	return nil
}

func previewJellyseerr(cfg *config.Config) *integrations.Jellyseerr {
	return integrations.NewJellyseerr(integrations.JellyseerrConfig{URL: cfg.Jellyseerr.URL, APIKey: cfg.Jellyseerr.APIKey, UserID: cfg.Jellyseerr.UserID, RequestEmail: cfg.Jellyseerr.RequestCredentials.Email, RequestPassword: cfg.Jellyseerr.RequestCredentials.Password})
}
