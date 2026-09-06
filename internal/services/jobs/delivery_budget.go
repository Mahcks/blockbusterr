package jobs

import (
	"log/slog"
	"time"

	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/database"
	"github.com/mahcks/blockbusterr/internal/integrations"
	"github.com/mahcks/blockbusterr/pkg/enums"
)

type deliveryBudget struct {
	config      *config.Config
	database    *database.Database
	jobID       string
	runID       int64
	perRunLimit int
	dryRun      bool
	used        int
}

func newDeliveryBudget(cfg *config.Config, db *database.Database, jobID string, runID int64, perRunLimit int, dryRun bool) *deliveryBudget {
	return &deliveryBudget{config: cfg, database: db, jobID: jobID, runID: runID, perRunLimit: perRunLimit, dryRun: dryRun}
}

func (b *deliveryBudget) reserve(mediaType string) (int64, bool, string) {
	if b.perRunLimit > 0 && b.used >= b.perRunLimit {
		return 0, false, "Job delivery limit reached"
	}

	globalLimit := b.config.Jobs.GlobalLimitMovies
	if mediaType == "show" {
		globalLimit = b.config.Jobs.GlobalLimitShows
	}
	if b.database == nil {
		if globalLimit > 0 {
			return 0, false, "Delivery budget unavailable"
		}
		b.used++
		return 0, true, ""
	}

	since := time.Now().Add(-deliveryBudgetWindow(b.config.Jobs.GlobalPeriod))
	if b.dryRun {
		if globalLimit > 0 {
			count, err := b.database.CountDeliveriesSince(mediaType, since)
			if err != nil {
				slog.Error("Failed to read delivery budget", "job_id", b.jobID, "media_type", mediaType, "err", err)
				return 0, false, "Delivery budget unavailable"
			}
			if count+b.used >= globalLimit {
				return 0, false, "Global delivery limit reached"
			}
		}
		b.used++
		return 0, true, ""
	}

	perRunLimit := b.perRunLimit
	if b.runID == 0 {
		perRunLimit = 0
	}
	id, reason, err := b.database.TryReserveDelivery(b.runID, b.jobID, mediaType, perRunLimit, globalLimit, since)
	if err != nil {
		slog.Error("Failed to reserve delivery budget", "job_id", b.jobID, "media_type", mediaType, "err", err)
		return 0, false, "Delivery budget unavailable"
	}
	if reason != "" {
		return 0, false, reason
	}
	b.used++
	return id, true, ""
}

func (b *deliveryBudget) release(id int64) {
	if b.used > 0 {
		b.used--
	}
	if b.database != nil && id != 0 {
		if err := b.database.ReleaseDeliveryReservation(id); err != nil {
			slog.Error("Failed to release delivery budget", "reservation_id", id, "err", err)
		}
	}
}

func deliveryBudgetWindow(period string) time.Duration {
	switch period {
	case "weekly":
		return 7 * 24 * time.Hour
	case "monthly":
		return 30 * 24 * time.Hour
	default:
		return 24 * time.Hour
	}
}

func (e *MovieJobExecutor) skipMovieDelivery(jobConfig JobConfig, movie integrations.Movie, scoreMap map[string]ScoreInfo, reason string) {
	e.updateDecisionOutcome(movie.IDs, "skipped", reason)
	if e.Database == nil {
		return
	}
	scoreInfo := scoreMap[integrations.MovieKey(movie.IDs)]
	if err := e.Database.LogActivity(database.ActivityLog{
		Timestamp: time.Now(), JobID: jobConfig.JobID, RunID: e.currentRunID, JobType: jobConfig.JobName,
		MediaType: "movie", Title: movie.Title, Year: movie.Year, Language: movie.Language,
		TMDBID: movie.IDs.TMDB, IMDBID: movie.IDs.IMDB, PosterURL: GetTMDBPosterURL(e.Config, movie.IDs.TMDB, "movie"),
		Score: scoreInfo.Score, Rank: scoreInfo.Rank, Status: "skipped", Message: reason,
		FilterDetails: e.getFilterDetailsForMovie(movie.IDs),
	}); err != nil {
		slog.Error("Failed to log delivery budget skip", "title", movie.Title, "err", err)
	}
}

func (e *ShowJobExecutor) skipShowDelivery(jobConfig JobConfig, show integrations.Show, scoreMap map[string]ScoreInfo, reason string) {
	e.updateDecisionOutcome(show.IDs, "skipped", reason)
	if e.Database == nil {
		return
	}
	scoreInfo := scoreMap[integrations.ShowKey(show.IDs)]
	if err := e.Database.LogActivity(database.ActivityLog{
		Timestamp: time.Now(), JobID: jobConfig.JobID, RunID: e.currentRunID, JobType: jobConfig.JobName,
		MediaType: "show", Title: show.Title, Year: show.Year, Language: show.Language,
		TMDBID: show.IDs.TMDB, TVDBID: show.IDs.TVDB, IMDBID: show.IDs.IMDB, PosterURL: GetShowPosterURL(e.Config, show.IDs.TMDB, show.IDs.TVDB),
		Score: scoreInfo.Score, Rank: scoreInfo.Rank, Status: "skipped", Message: reason,
		FilterDetails: e.getFilterDetailsForShow(show.IDs),
	}); err != nil {
		slog.Error("Failed to log delivery budget skip", "title", show.Title, "err", err)
	}
}

func (e *ShowJobExecutor) failShowDelivery(jobConfig JobConfig, show integrations.Show, scoreMap map[string]ScoreInfo, reason string) {
	e.updateDecisionOutcome(show.IDs, string(enums.ActivityStatusFailed), reason)
	if e.Database == nil {
		return
	}
	scoreInfo := scoreMap[integrations.ShowKey(show.IDs)]
	if err := e.Database.LogActivity(database.ActivityLog{
		Timestamp: time.Now(), JobID: jobConfig.JobID, RunID: e.currentRunID, JobType: jobConfig.JobName,
		MediaType: "show", Title: show.Title, Year: show.Year, Language: show.Language,
		TMDBID: show.IDs.TMDB, TVDBID: show.IDs.TVDB, IMDBID: show.IDs.IMDB, PosterURL: GetShowPosterURL(e.Config, show.IDs.TMDB, show.IDs.TVDB),
		Score: scoreInfo.Score, Rank: scoreInfo.Rank, Status: string(enums.ActivityStatusFailed), Message: reason,
		FilterDetails: e.getFilterDetailsForShow(show.IDs),
	}); err != nil {
		slog.Error("Failed to log show delivery failure", "title", show.Title, "err", err)
	}
}
