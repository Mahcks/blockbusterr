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

type jobConfig struct {
	name         string
	enabled      bool
	syncInterval string
	mode         string
	runFunc      func()
}

type Scheduler struct {
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	getConfig func() *config.Config
	db        *database.Database
	version   string
	jobStops  map[string]context.CancelFunc
	jobMutex  sync.Mutex
}

func NewScheduler(getConfig func() *config.Config, db *database.Database, version string) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		ctx:       ctx,
		cancel:    cancel,
		getConfig: getConfig,
		db:        db,
		version:   version,
		jobStops:  make(map[string]context.CancelFunc),
	}
}

func (s *Scheduler) Start() {
	log.Info("Starting job scheduler")

	s.wg.Go(func() {
		s.run()
	})
}

func (s *Scheduler) Stop() {
	log.Info("Stopping job scheduler")
	s.cancel()

	// Stop all individual job goroutines
	s.jobMutex.Lock()
	for name, cancel := range s.jobStops {
		log.Infof("Stopping job: %s", name)
		cancel()
	}
	s.jobMutex.Unlock()

	s.wg.Wait()
	log.Info("Job scheduler stopped")
}

func (s *Scheduler) run() {
	// Initial execution of all enabled jobs
	s.executeAllJobs()

	// Start individual job schedulers
	s.scheduleJobs()

	// Monitor for config changes
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			// Re-schedule jobs if config changed
			s.scheduleJobs()
		}
	}
}

func (s *Scheduler) scheduleJobs() {
	cfg := s.getConfig()
	dryRun := s.version == "dev"
	defaultInterval := cfg.Jobs.SyncInterval
	defaultMode := cfg.Jobs.Mode
	if defaultMode == "" {
		defaultMode = "direct"
	}

	// Build job configurations
	jobConfigs := []jobConfig{
		{
			name:         "trending_movies",
			enabled:      cfg.Jobs.TrendingMovies.Enabled,
			syncInterval: getJobInterval(cfg.Jobs.TrendingMovies.SyncInterval, defaultInterval),
			mode:         getJobMode(cfg.Jobs.TrendingMovies.Mode, defaultMode),
			runFunc:      func() { jobs.RunTrendingMovies(cfg, s.db, dryRun) },
		},
		{
			name:         "trending_shows",
			enabled:      cfg.Jobs.TrendingShows.Enabled,
			syncInterval: getJobInterval(cfg.Jobs.TrendingShows.SyncInterval, defaultInterval),
			mode:         getJobMode(cfg.Jobs.TrendingShows.Mode, defaultMode),
			runFunc:      func() { jobs.RunTrendingShows(cfg, s.db, dryRun) },
		},
		{
			name:         "popular_movies",
			enabled:      cfg.Jobs.PopularMovies.Enabled,
			syncInterval: getJobInterval(cfg.Jobs.PopularMovies.SyncInterval, defaultInterval),
			mode:         getJobMode(cfg.Jobs.PopularMovies.Mode, defaultMode),
			runFunc:      func() { jobs.RunPopularMovies(cfg, s.db, dryRun) },
		},
		{
			name:         "popular_shows",
			enabled:      cfg.Jobs.PopularShows.Enabled,
			syncInterval: getJobInterval(cfg.Jobs.PopularShows.SyncInterval, defaultInterval),
			mode:         getJobMode(cfg.Jobs.PopularShows.Mode, defaultMode),
			runFunc:      func() { jobs.RunPopularShows(cfg, s.db, dryRun) },
		},
		{
			name:         "box_office",
			enabled:      cfg.Jobs.BoxOffice.Enabled,
			syncInterval: getJobInterval(cfg.Jobs.BoxOffice.SyncInterval, defaultInterval),
			mode:         getJobMode(cfg.Jobs.BoxOffice.Mode, defaultMode),
			runFunc:      func() { jobs.RunBoxOffice(cfg, s.db, dryRun) },
		},
		{
			name:         "favorited_movies",
			enabled:      cfg.Jobs.FavoritedMovies.Enabled,
			syncInterval: getJobInterval(cfg.Jobs.FavoritedMovies.SyncInterval, defaultInterval),
			mode:         getJobMode(cfg.Jobs.FavoritedMovies.Mode, defaultMode),
			runFunc:      func() { jobs.RunFavoritedMovies(cfg, s.db, dryRun) },
		},
		{
			name:         "played_movies",
			enabled:      cfg.Jobs.PlayedMovies.Enabled,
			syncInterval: getJobInterval(cfg.Jobs.PlayedMovies.SyncInterval, defaultInterval),
			mode:         getJobMode(cfg.Jobs.PlayedMovies.Mode, defaultMode),
			runFunc:      func() { jobs.RunPlayedMovies(cfg, s.db, dryRun) },
		},
		{
			name:         "watched_movies",
			enabled:      cfg.Jobs.WatchedMovies.Enabled,
			syncInterval: getJobInterval(cfg.Jobs.WatchedMovies.SyncInterval, defaultInterval),
			mode:         getJobMode(cfg.Jobs.WatchedMovies.Mode, defaultMode),
			runFunc:      func() { jobs.RunWatchedMovies(cfg, s.db, dryRun) },
		},
		{
			name:         "collected_movies",
			enabled:      cfg.Jobs.CollectedMovies.Enabled,
			syncInterval: getJobInterval(cfg.Jobs.CollectedMovies.SyncInterval, defaultInterval),
			mode:         getJobMode(cfg.Jobs.CollectedMovies.Mode, defaultMode),
			runFunc:      func() { jobs.RunCollectedMovies(cfg, s.db, dryRun) },
		},
		{
			name:         "anticipated_movies",
			enabled:      cfg.Jobs.AnticipatedMovies.Enabled,
			syncInterval: getJobInterval(cfg.Jobs.AnticipatedMovies.SyncInterval, defaultInterval),
			mode:         getJobMode(cfg.Jobs.AnticipatedMovies.Mode, defaultMode),
			runFunc:      func() { jobs.RunAnticipatedMovies(cfg, s.db, dryRun) },
		},
		{
			name:         "favorited_shows",
			enabled:      cfg.Jobs.FavoritedShows.Enabled,
			syncInterval: getJobInterval(cfg.Jobs.FavoritedShows.SyncInterval, defaultInterval),
			mode:         getJobMode(cfg.Jobs.FavoritedShows.Mode, defaultMode),
			runFunc:      func() { jobs.RunFavoritedShows(cfg, s.db, dryRun) },
		},
		{
			name:         "played_shows",
			enabled:      cfg.Jobs.PlayedShows.Enabled,
			syncInterval: getJobInterval(cfg.Jobs.PlayedShows.SyncInterval, defaultInterval),
			mode:         getJobMode(cfg.Jobs.PlayedShows.Mode, defaultMode),
			runFunc:      func() { jobs.RunPlayedShows(cfg, s.db, dryRun) },
		},
		{
			name:         "watched_shows",
			enabled:      cfg.Jobs.WatchedShows.Enabled,
			syncInterval: getJobInterval(cfg.Jobs.WatchedShows.SyncInterval, defaultInterval),
			mode:         getJobMode(cfg.Jobs.WatchedShows.Mode, defaultMode),
			runFunc:      func() { jobs.RunWatchedShows(cfg, s.db, dryRun) },
		},
		{
			name:         "collected_shows",
			enabled:      cfg.Jobs.CollectedShows.Enabled,
			syncInterval: getJobInterval(cfg.Jobs.CollectedShows.SyncInterval, defaultInterval),
			mode:         getJobMode(cfg.Jobs.CollectedShows.Mode, defaultMode),
			runFunc:      func() { jobs.RunCollectedShows(cfg, s.db, dryRun) },
		},
		{
			name:         "anticipated_shows",
			enabled:      cfg.Jobs.AnticipatedShows.Enabled,
			syncInterval: getJobInterval(cfg.Jobs.AnticipatedShows.SyncInterval, defaultInterval),
			mode:         getJobMode(cfg.Jobs.AnticipatedShows.Mode, defaultMode),
			runFunc:      func() { jobs.RunAnticipatedShows(cfg, s.db, dryRun) },
		},
		{
			name:         "smart_popular_movies",
			enabled:      cfg.Jobs.SmartPopularMovies.Enabled,
			syncInterval: getJobInterval(cfg.Jobs.SmartPopularMovies.SyncInterval, defaultInterval),
			mode:         getJobMode(cfg.Jobs.SmartPopularMovies.Mode, defaultMode),
			runFunc:      func() { jobs.RunSmartPopularMovies(cfg, s.db, dryRun) },
		},
		{
			name:         "smart_popular_shows",
			enabled:      cfg.Jobs.SmartPopularShows.Enabled,
			syncInterval: getJobInterval(cfg.Jobs.SmartPopularShows.SyncInterval, defaultInterval),
			mode:         getJobMode(cfg.Jobs.SmartPopularShows.Mode, defaultMode),
			runFunc:      func() { jobs.RunSmartPopularShows(cfg, s.db, dryRun) },
		},
	}

	s.jobMutex.Lock()
	defer s.jobMutex.Unlock()

	// Schedule or reschedule each job
	for _, jc := range jobConfigs {
		if !jc.enabled {
			// Stop job if it's running but now disabled
			if cancel, exists := s.jobStops[jc.name]; exists {
				cancel()
				delete(s.jobStops, jc.name)
			}
			continue
		}

		// Check if job is already scheduled
		if _, exists := s.jobStops[jc.name]; exists {
			// Job already running, skip (we could add logic to detect config changes and restart)
			continue
		}

		// Start new job scheduler
		s.startJobScheduler(jc)
	}
}

func (s *Scheduler) startJobScheduler(jc jobConfig) {
	jobCtx, jobCancel := context.WithCancel(s.ctx)
	s.jobStops[jc.name] = jobCancel

	s.wg.Add(1)
	go func(jc jobConfig) {
		defer s.wg.Done()
		defer func() {
			s.jobMutex.Lock()
			delete(s.jobStops, jc.name)
			s.jobMutex.Unlock()
		}()

		duration, cronSchedule, _ := parseSyncInterval(jc.syncInterval)

		// Log job startup
		if cronSchedule != nil {
			nextRun := (*cronSchedule).Next(time.Now())
			log.Infof("Job '%s' scheduled with cron '%s' (%s mode), next run at %s",
				jc.name, jc.syncInterval, jc.mode, nextRun.Format(time.RFC3339))
		} else {
			if duration == 0 {
				log.Warnf("Invalid sync interval '%s' for job '%s', defaulting to 1h", jc.syncInterval, jc.name)
				duration = 1 * time.Hour
			}
			log.Infof("Job '%s' scheduled every %s (%s mode)", jc.name, duration, jc.mode)
		}

		// Create ticker
		var ticker *time.Ticker
		if cronSchedule != nil {
			nextRun := (*cronSchedule).Next(time.Now())
			waitDuration := time.Until(nextRun)
			ticker = time.NewTicker(waitDuration)
		} else {
			ticker = time.NewTicker(duration)
		}
		defer ticker.Stop()

		for {
			select {
			case <-jobCtx.Done():
				log.Infof("Job '%s' stopped", jc.name)
				return
			case <-ticker.C:
				// Execute the job
				jc.runFunc()

				// If using cron, calculate next run time
				if cronSchedule != nil {
					nextRun := (*cronSchedule).Next(time.Now())
					waitDuration := time.Until(nextRun)
					ticker.Reset(waitDuration)
				}
			}
		}
	}(jc)
}

func (s *Scheduler) executeAllJobs() {
	cfg := s.getConfig()
	dryRun := s.version == "dev"

	if dryRun {
		log.Info("Executing initial jobs in DRY RUN mode (no content will be added)")
	} else {
		log.Info("Executing initial jobs")
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

	log.Info("Completed initial job execution")
}

// Helper functions
func getJobInterval(jobInterval, defaultInterval string) string {
	if jobInterval != "" {
		return jobInterval
	}
	if defaultInterval != "" {
		return defaultInterval
	}
	return "1h" // Ultimate fallback
}

func getJobMode(jobMode, defaultMode string) string {
	if jobMode != "" {
		return jobMode
	}
	return defaultMode
}

func parseSyncInterval(interval string) (time.Duration, *cron.Schedule, error) {
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
