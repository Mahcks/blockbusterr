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
	"github.com/mahcks/blockbusterr/pkg/enums"
)

// PreviewDynamicJob uses the same provider selection and fetchers as job execution.
func PreviewDynamicJob(cfg *config.Config, db *database.Database, job config.DynamicJob) (PreviewResponse, error) {
	return previewDynamicJob(cfg, db, job, nil)
}

func previewDynamicJob(cfg *config.Config, db *database.Database, job config.DynamicJob, listSources ListSourceRegistry) (PreviewResponse, error) {
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
	job.Mode = mode
	if job.Type == "smart_popular" {
		if job.BaseMinRating == 0 {
			job.BaseMinRating = 6
		}
		if job.AdjustmentFactor == 0 {
			job.AdjustmentFactor = .5
		}
	}
	response := PreviewResponse{JobName: job.Name, Source: job.Source, MediaType: job.MediaType, Mode: mode, HasPosters: hasTMDBConfigured(cfg), Items: []PreviewItem{}}
	executor := &DynamicJobExecutor{Config: cfg, Database: db, DryRun: true, ListSources: listSources}
	discovery, err := executor.discoveryForJob(job)
	if err != nil {
		return response, err
	}

	if job.MediaType == "movie" {
		fetcher := executor.getMovieFetcher(job)
		if fetcher == nil {
			return response, fmt.Errorf("no movie fetcher for job type: %s", job.Type)
		}
		movies, err := fetcher(ctx, discovery, job.Limit, job.Period)
		if err != nil {
			return response, err
		}
		response.TotalFound = len(movies)
		if err := previewMovies(ctx, cfg, db, job, movies, &response); err != nil {
			return response, err
		}
		return response, nil
	}

	fetcher := executor.getShowFetcher(job)
	if fetcher == nil {
		return response, fmt.Errorf("no show fetcher for job type: %s", job.Type)
	}
	shows, err := fetcher(ctx, discovery, job.Limit, job.Period)
	if err != nil {
		return response, err
	}
	response.TotalFound = len(shows)
	if err := previewShows(ctx, cfg, db, job, shows, &response); err != nil {
		return response, err
	}
	return response, nil
}

func previewMovies(ctx context.Context, cfg *config.Config, db *database.Database, job config.DynamicJob, movies []integrations.Movie, response *PreviewResponse) error {
	mode, repeatPolicy := job.Mode, job.RepeatPolicy
	var percentiles map[int]float64
	if job.Type == "smart_popular" {
		percentiles = filters.CalculateMoviePopularityPercentiles(movies)
	}
	if err := enrichMovieCertifications(ctx, cfg, movies); err != nil {
		return fmt.Errorf("failed to enrich movie certifications: %w", err)
	}
	scores := ScoreAndRankMovies(movies, cfg)
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
	deliveries, err := previewDeliveryHistory(db, movieDeliveryIdentities(movies))
	if err != nil {
		return err
	}
	policy := effectiveRepeatPolicy(repeatPolicy, cfg.Jobs.RepeatPolicy)
	for index, movie := range movies {
		item := createMoviePreviewItem(cfg, movie, len(movies)-index)
		score := scores[integrations.MovieKey(movie.IDs)]
		item.Score, item.Rank = score.Score, score.Rank
		item.ProviderRank = index + 1
		result := filters.MoviePassesRules(movie, cfg.Filters.Movies, cfg.TitleExceptions)
		if percentiles != nil {
			threshold := filters.CalculateAdaptiveRating(job.BaseMinRating, percentiles[movie.IDs.TMDB], job.AdjustmentFactor)
			result = filters.MoviePassesAdaptiveFilters(movie, cfg.Filters.Movies, threshold, cfg.TitleExceptions)
		}
		item.FilterChecks = result.Checks
		item.DecisionReason = filters.Explain(result)
		if !result.Passed {
			item.FilteredOut, item.FilterReason = true, item.DecisionReason
			response.FilteredOut++
		} else if reason := previewRepeatReason(policy, deliveries, database.DeliveryIdentity{MediaType: "movie", TMDBID: movie.IDs.TMDB, IMDBID: movie.IDs.IMDB}, time.Now()); reason != "" {
			item.AlreadyExists, item.RepeatBlocked, item.DecisionReason = true, true, "Skipped: "+reason
			response.AlreadyExists++
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
				item.DecisionReason = previewDeliveryReason(mode, result)
				response.WillAdd++
			}
		} else {
			item.DecisionReason = previewDeliveryReason(mode, result)
			response.WillAdd++
		}
		response.Items = append(response.Items, item)
	}
	return nil
}

func previewShows(ctx context.Context, cfg *config.Config, db *database.Database, job config.DynamicJob, shows []integrations.Show, response *PreviewResponse) error {
	mode, repeatPolicy := job.Mode, job.RepeatPolicy
	var percentiles map[int]float64
	if job.Type == "smart_popular" {
		percentiles = filters.CalculateShowPopularityPercentiles(shows)
	}
	if err := enrichShowCertifications(ctx, cfg, shows); err != nil {
		return fmt.Errorf("failed to enrich show certifications: %w", err)
	}
	scores := ScoreAndRankShows(shows, cfg)
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
	deliveries, err := previewDeliveryHistory(db, showDeliveryIdentities(shows))
	if err != nil {
		return err
	}
	policy := effectiveRepeatPolicy(repeatPolicy, cfg.Jobs.RepeatPolicy)
	for index, show := range shows {
		item := createShowPreviewItem(cfg, show, len(shows)-index)
		score := scores[integrations.ShowKey(show.IDs)]
		item.Score, item.Rank = score.Score, score.Rank
		item.ProviderRank = index + 1
		result := filters.ShowPassesRules(show, cfg.Filters.Shows, cfg.TitleExceptions)
		if percentiles != nil {
			threshold := filters.CalculateAdaptiveRating(job.BaseMinRating, percentiles[show.IDs.TVDB], job.AdjustmentFactor)
			result = filters.ShowPassesAdaptiveFilters(show, cfg.Filters.Shows, threshold, cfg.TitleExceptions)
		}
		item.FilterChecks = result.Checks
		item.DecisionReason = filters.Explain(result)
		if !result.Passed {
			item.FilteredOut, item.FilterReason = true, item.DecisionReason
			response.FilteredOut++
		} else if reason := previewRepeatReason(policy, deliveries, database.DeliveryIdentity{MediaType: "show", TMDBID: show.IDs.TMDB, TVDBID: show.IDs.TVDB, IMDBID: show.IDs.IMDB}, time.Now()); reason != "" {
			item.AlreadyExists, item.RepeatBlocked, item.DecisionReason = true, true, "Skipped: "+reason
			response.AlreadyExists++
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
				item.DecisionReason = previewDeliveryReason(mode, result)
				response.WillAdd++
			}
		} else {
			item.DecisionReason = previewDeliveryReason(mode, result)
			response.WillAdd++
		}
		response.Items = append(response.Items, item)
	}
	return nil
}

func previewDeliveryReason(mode string, result filters.FilterResult) string {
	verb := "add"
	if mode == "jellyseerr" {
		verb = "request"
	}
	return "Will " + verb + ": " + strings.TrimPrefix(filters.Explain(result), "Accepted: ")
}

func previewDeliveryHistory(db *database.Database, identities []database.DeliveryIdentity) (map[string]time.Time, error) {
	if db == nil {
		return map[string]time.Time{}, nil
	}
	return db.LatestSuccessfulDeliveries(identities)
}

func previewRepeatReason(policy enums.RepeatPolicy, deliveries map[string]time.Time, identity database.DeliveryIdentity, now time.Time) string {
	if policy == enums.RepeatPolicyImmediate {
		return ""
	}
	deliveredAt, ok := deliveries[identity.Key()]
	if !ok {
		return ""
	}
	return repeatSkipReasonForDelivery(policy, deliveredAt, now)
}

func movieDeliveryIdentities(movies []integrations.Movie) []database.DeliveryIdentity {
	identities := make([]database.DeliveryIdentity, 0, len(movies))
	for _, movie := range movies {
		identities = append(identities, database.DeliveryIdentity{MediaType: "movie", TMDBID: movie.IDs.TMDB, IMDBID: movie.IDs.IMDB})
	}
	return identities
}

func showDeliveryIdentities(shows []integrations.Show) []database.DeliveryIdentity {
	identities := make([]database.DeliveryIdentity, 0, len(shows))
	for _, show := range shows {
		identities = append(identities, database.DeliveryIdentity{MediaType: "show", TMDBID: show.IDs.TMDB, TVDBID: show.IDs.TVDB, IMDBID: show.IDs.IMDB})
	}
	return identities
}

func previewJellyseerr(cfg *config.Config) *integrations.Jellyseerr {
	return integrations.NewJellyseerr(integrations.JellyseerrConfig{URL: cfg.Jellyseerr.URL, APIKey: cfg.Jellyseerr.APIKey, UserID: cfg.Jellyseerr.UserID, RequestEmail: cfg.Jellyseerr.RequestCredentials.Email, RequestPassword: cfg.Jellyseerr.RequestCredentials.Password})
}
