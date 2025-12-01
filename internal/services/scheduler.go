package services

import (
	"context"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/database"
	"github.com/mahcks/blockbusterr/internal/services/jobs"
	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	getConfig func() *config.Config
	db        *database.Database
	version   string
}

func NewScheduler(getConfig func() *config.Config, db *database.Database, version string) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		ctx:       ctx,
		cancel:    cancel,
		getConfig: getConfig,
		db:        db,
		version:   version,
	}
}

func (s *Scheduler) Start() {
	log.Info("Starting job scheduler")

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.run()
	}()
}

func (s *Scheduler) Stop() {
	log.Info("Stopping job scheduler")
	s.cancel()
	s.wg.Wait()
	log.Info("Job scheduler stopped")
}

func (s *Scheduler) run() {
	// Run jobs immediately on startup
	s.executeJobs()

	cfg := s.getConfig()

	// Helper to parse sync interval - supports both duration (e.g. "1h") and cron expressions (e.g. "0 */2 * * *")
	parseSyncInterval := func(interval string) (time.Duration, *cron.Schedule, error) {
		// Try parsing as duration first
		if duration, err := time.ParseDuration(interval); err == nil {
			return duration, nil, nil
		}

		// Try parsing as cron expression
		parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
		if schedule, err := parser.Parse(interval); err == nil {
			return 0, &schedule, nil
		}

		return 0, nil, nil
	}

	duration, cronSchedule, _ := parseSyncInterval(cfg.Jobs.SyncInterval)
	currentSyncInterval := cfg.Jobs.SyncInterval

	// If it's a cron schedule, calculate next run time
	var ticker *time.Ticker
	if cronSchedule != nil {
		nextRun := (*cronSchedule).Next(time.Now())
		waitDuration := time.Until(nextRun)
		log.Infof("Jobs will run on cron schedule '%s', next run at %s", cfg.Jobs.SyncInterval, nextRun.Format(time.RFC3339))
		ticker = time.NewTicker(waitDuration)
	} else {
		if duration == 0 {
			log.Warnf("Invalid sync interval '%s', defaulting to 1h", cfg.Jobs.SyncInterval)
			duration = 1 * time.Hour
		}
		log.Infof("Jobs will run every %s", duration.String())
		ticker = time.NewTicker(duration)
	}
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			// Reload config to get updated sync interval
			cfg := s.getConfig()
			newDuration, newCronSchedule, _ := parseSyncInterval(cfg.Jobs.SyncInterval)

			intervalChanged := false
			if newCronSchedule != nil && cronSchedule != nil {
				// Both cron - check if expression changed
				intervalChanged = cfg.Jobs.SyncInterval != currentSyncInterval
			} else if newCronSchedule == nil && cronSchedule == nil {
				// Both duration - check if duration changed
				intervalChanged = newDuration != duration
			} else {
				// Type changed (duration<->cron)
				intervalChanged = true
			}
			if intervalChanged {
				duration = newDuration
				cronSchedule = newCronSchedule
				currentSyncInterval = cfg.Jobs.SyncInterval

				// Reset ticker
				// Reset ticker
				if cronSchedule != nil {
					nextRun := (*cronSchedule).Next(time.Now())
					waitDuration := time.Until(nextRun)
					ticker.Reset(waitDuration)
					log.Infof("Updated job schedule to cron '%s', next run at %s", cfg.Jobs.SyncInterval, nextRun.Format(time.RFC3339))
				} else {
					ticker.Reset(duration)
					log.Infof("Updated job sync interval to %s", duration.String())
				}
			}

			// Execute jobs
			s.executeJobs()

			// If using cron, calculate next run time after executing
			if cronSchedule != nil {
				nextRun := (*cronSchedule).Next(time.Now())
				waitDuration := time.Until(nextRun)
				ticker.Reset(waitDuration)
			}
		}
	}
}

func (s *Scheduler) executeJobs() {
	cfg := s.getConfig()
	dryRun := s.version == "dev"

	if dryRun {
		log.Info("Executing scheduled jobs in DRY RUN mode (no content will be added)")
	} else {
		log.Info("Executing scheduled jobs")
	}

	// Run each enabled job
	if cfg.Jobs.TrendingMovies.Enabled {
		jobs.RunTrendingMovies(cfg, s.db, dryRun)
	}

	if cfg.Jobs.TrendingShows.Enabled {
		jobs.RunTrendingShows(cfg, s.db, dryRun)
	}

	if cfg.Jobs.PopularMovies.Enabled {
		jobs.RunPopularMovies(cfg, s.db, dryRun)
	}

	if cfg.Jobs.PopularShows.Enabled {
		jobs.RunPopularShows(cfg, s.db, dryRun)
	}

	if cfg.Jobs.BoxOffice.Enabled {
		jobs.RunBoxOffice(cfg, s.db, dryRun)
	}

	if cfg.Jobs.FavoritedMovies.Enabled {
		jobs.RunFavoritedMovies(cfg, s.db, dryRun)
	}

	if cfg.Jobs.PlayedMovies.Enabled {
		jobs.RunPlayedMovies(cfg, s.db, dryRun)
	}

	if cfg.Jobs.WatchedMovies.Enabled {
		jobs.RunWatchedMovies(cfg, s.db, dryRun)
	}

	if cfg.Jobs.CollectedMovies.Enabled {
		jobs.RunCollectedMovies(cfg, s.db, dryRun)
	}

	if cfg.Jobs.AnticipatedMovies.Enabled {
		jobs.RunAnticipatedMovies(cfg, s.db, dryRun)
	}

	if cfg.Jobs.FavoritedShows.Enabled {
		jobs.RunFavoritedShows(cfg, s.db, dryRun)
	}

	if cfg.Jobs.PlayedShows.Enabled {
		jobs.RunPlayedShows(cfg, s.db, dryRun)
	}

	if cfg.Jobs.WatchedShows.Enabled {
		jobs.RunWatchedShows(cfg, s.db, dryRun)
	}

	if cfg.Jobs.CollectedShows.Enabled {
		jobs.RunCollectedShows(cfg, s.db, dryRun)
	}

	if cfg.Jobs.AnticipatedShows.Enabled {
		jobs.RunAnticipatedShows(cfg, s.db, dryRun)
	}

	log.Info("Completed job execution cycle")
}
