package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/database"
	"github.com/mahcks/blockbusterr/internal/services/jobs"
	"github.com/robfig/cron/v3"
)

type jobConfig struct {
	id           string
	name         string
	enabled      bool
	syncInterval string
	mode         string
	runFunc      func(context.Context)
}

type Scheduler struct {
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	getConfig      func() *config.Config
	db             *database.Database
	version        string
	executions     *jobs.ExecutionCoordinator
	jobStops       map[string]context.CancelFunc
	jobSigs        map[string]string
	jobNames       map[string]string
	jobGenerations map[string]uint64
	nextGeneration uint64
	jobMutex       sync.Mutex
}

func NewScheduler(getConfig func() *config.Config, db *database.Database, version string, executions *jobs.ExecutionCoordinator) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		ctx:            ctx,
		cancel:         cancel,
		getConfig:      getConfig,
		db:             db,
		version:        version,
		executions:     executions,
		jobStops:       make(map[string]context.CancelFunc),
		jobSigs:        make(map[string]string),
		jobNames:       make(map[string]string),
		jobGenerations: make(map[string]uint64),
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
	for id, cancel := range s.jobStops {
		log.Infof("Stopping job: %s", formatJobLabel(s.jobNames[id], id))
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

	// Get all enabled jobs (both dynamic and legacy)
	enabledJobs := runnableJobs(cfg)
	cycleJobs, enabledJobs := splitSelectionJobs(cfg, enabledJobs)

	// Track which jobs should be running
	activeJobIDs := make(map[string]bool)
	for _, job := range enabledJobs {
		activeJobIDs[job.ID] = true
	}
	if len(cycleJobs) > 0 {
		activeJobIDs[selectionCycleJobID] = true
	}

	s.jobMutex.Lock()
	defer s.jobMutex.Unlock()

	// Stop jobs that are no longer enabled
	for jobID, cancel := range s.jobStops {
		if !activeJobIDs[jobID] {
			log.Infof("Stopping disabled job: %s", formatJobLabel(s.jobNames[jobID], jobID))
			cancel()
			delete(s.jobStops, jobID)
			delete(s.jobSigs, jobID)
			delete(s.jobNames, jobID)
			delete(s.jobGenerations, jobID)
		}
	}

	// Schedule each enabled job
	for _, job := range enabledJobs {
		syncInterval := getJobInterval(job.SyncInterval, defaultInterval)
		mode := getJobMode(job.Mode, cfg.Jobs.Mode)
		jobSig := jobSignature(job) + "|" + syncInterval + "|" + mode

		// Check if job is already scheduled
		if _, exists := s.jobStops[job.ID]; exists {
			// If the signature changed, restart the job to apply updates
			if s.jobSigs[job.ID] != jobSig {
				log.Infof("Job '%s' changed, rescheduling", formatJobLabel(job.Name, job.ID))
				s.jobStops[job.ID]()
				delete(s.jobStops, job.ID)
				delete(s.jobSigs, job.ID)
				delete(s.jobNames, job.ID)
				delete(s.jobGenerations, job.ID)
			} else {
				// Job already running with same config, skip
				continue
			}
		}

		// Create job config for the scheduler
		jc := jobConfig{
			id:           job.ID,
			name:         job.Name,
			enabled:      job.Enabled,
			syncInterval: syncInterval,
			mode:         mode,
			runFunc: func(jobID, jobName string) func(context.Context) {
				return func(ctx context.Context) {
					cfg := s.getConfig()
					job := cfg.GetDynamicJobByID(jobID)
					if job == nil || !job.Enabled {
						return
					}
					if err := s.executions.RunDynamicJob(ctx, cfg, s.db, *job, dryRun); err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, jobs.ErrExecutionAlreadyRunning) {
						log.Errorf("Failed to run job %s: %v", formatJobLabel(jobName, jobID), err)
					}
				}
			}(job.ID, job.Name),
		}

		// Start new job scheduler
		s.startJobScheduler(jc)
		s.jobSigs[job.ID] = jobSig
		s.jobNames[job.ID] = job.Name
	}

	if len(cycleJobs) > 0 {
		sig := selectionCycleSignature(cfg, cycleJobs)
		if _, exists := s.jobStops[selectionCycleJobID]; exists && s.jobSigs[selectionCycleJobID] != sig {
			s.jobStops[selectionCycleJobID]()
			delete(s.jobStops, selectionCycleJobID)
			delete(s.jobSigs, selectionCycleJobID)
			delete(s.jobNames, selectionCycleJobID)
			delete(s.jobGenerations, selectionCycleJobID)
		}
		if _, exists := s.jobStops[selectionCycleJobID]; !exists {
			s.startJobScheduler(jobConfig{id: selectionCycleJobID, name: "Ranked selection", enabled: true, syncInterval: cfg.Jobs.Selection.SyncInterval, mode: "shared", runFunc: func(ctx context.Context) {
				if _, err := s.executions.RunSelectionCycle(ctx, s.getConfig(), s.db, dryRun); err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, jobs.ErrExecutionAlreadyRunning) {
					log.Errorf("Failed to run ranked selection: %v", err)
				}
			}})
			s.jobSigs[selectionCycleJobID], s.jobNames[selectionCycleJobID] = sig, "Ranked selection"
		}
	}
}

func (s *Scheduler) startJobScheduler(jc jobConfig) {
	jobCtx, jobCancel := context.WithCancel(s.ctx)
	s.jobStops[jc.id] = jobCancel
	s.nextGeneration++
	generation := s.nextGeneration
	s.jobGenerations[jc.id] = generation

	s.wg.Add(1)
	go func(jc jobConfig) {
		defer s.wg.Done()
		defer s.clearJobSchedule(jc.id, generation)

		duration, cronSchedule, _ := parseSyncInterval(jc.syncInterval)

		// Log job startup
		if cronSchedule != nil {
			nextRun := (*cronSchedule).Next(time.Now())
			log.Infof("Job '%s' scheduled with cron '%s' (%s mode), next run at %s",
				formatJobLabel(jc.name, jc.id), jc.syncInterval, jc.mode, nextRun.Format(time.RFC3339))
		} else {
			if duration == 0 {
				log.Warnf("Invalid sync interval '%s' for job '%s', defaulting to 1h", jc.syncInterval, formatJobLabel(jc.name, jc.id))
				duration = 1 * time.Hour
			}
			log.Infof("Job '%s' scheduled every %s (%s mode)", formatJobLabel(jc.name, jc.id), duration, jc.mode)
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
				log.Infof("Job '%s' stopped", formatJobLabel(jc.name, jc.id))
				return
			case <-ticker.C:
				// Execute the job
				jc.runFunc(jobCtx)

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

func (s *Scheduler) clearJobSchedule(id string, generation uint64) {
	s.jobMutex.Lock()
	defer s.jobMutex.Unlock()
	if s.jobGenerations[id] != generation {
		return
	}
	delete(s.jobStops, id)
	delete(s.jobSigs, id)
	delete(s.jobNames, id)
	delete(s.jobGenerations, id)
}

func (s *Scheduler) executeAllJobs() {
	cfg := s.getConfig()
	dryRun := s.version == "dev"

	if dryRun {
		log.Info("Executing initial jobs in DRY RUN mode (no content will be added)")
	} else {
		log.Info("Executing initial jobs")
	}

	// Get all enabled jobs (both dynamic and legacy)
	enabledJobs := runnableJobs(cfg)
	cycleJobs, enabledJobs := splitSelectionJobs(cfg, enabledJobs)

	log.Infof("Found %d enabled jobs to execute", len(enabledJobs)+len(cycleJobs))
	if len(cycleJobs) > 0 && runsAtStartup(cfg.Jobs.Selection.SyncInterval) {
		if _, err := s.executions.RunSelectionCycle(s.ctx, cfg, s.db, dryRun); err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, jobs.ErrExecutionAlreadyRunning) {
			log.Errorf("Failed to run ranked selection: %v", err)
		}
	}

	// Run each enabled job
	for _, job := range enabledJobs {
		if !runsAtStartup(getJobInterval(job.SyncInterval, cfg.Jobs.SyncInterval)) {
			continue
		}
		log.Infof("Executing job: %s (%s %s)", formatJobLabel(job.Name, job.ID), job.Type, job.MediaType)
		if err := s.executions.RunDynamicJob(s.ctx, cfg, s.db, job, dryRun); err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, jobs.ErrExecutionAlreadyRunning) {
			log.Errorf("Failed to run job %s: %v", formatJobLabel(job.Name, job.ID), err)
		}
	}

	log.Info("Completed initial job execution")
}

const selectionCycleJobID = jobs.SelectionCycleExecutionID

func splitSelectionJobs(cfg *config.Config, enabled []config.DynamicJob) (cycle, standalone []config.DynamicJob) {
	for _, job := range enabled {
		if cfg.Jobs.Selection.Enabled && job.SelectionCycle {
			cycle = append(cycle, job)
		} else {
			standalone = append(standalone, job)
		}
	}
	return cycle, standalone
}

func selectionCycleSignature(cfg *config.Config, cycleJobs []config.DynamicJob) string {
	parts := []string{fmt.Sprintf("%t|%s|%d|%d", cfg.Jobs.Selection.Enabled, cfg.Jobs.Selection.SyncInterval, cfg.Jobs.Selection.MovieLimit, cfg.Jobs.Selection.ShowLimit)}
	for _, job := range cycleJobs {
		parts = append(parts, jobSignature(job))
	}
	return strings.Join(parts, "|")
}

func runnableJobs(cfg *config.Config) []config.DynamicJob {
	enabled := cfg.GetEnabledJobs()
	runnable := make([]config.DynamicJob, 0, len(enabled))
	for _, job := range enabled {
		if jobs.IsProviderConfigured(cfg, job.Source) {
			runnable = append(runnable, job)
		}
	}
	return runnable
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

func runsAtStartup(interval string) bool {
	_, cronSchedule, _ := parseSyncInterval(interval)
	return cronSchedule == nil
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

func jobSignature(job config.DynamicJob) string {
	data, _ := json.Marshal(job)
	return string(data)
}

func formatJobLabel(name, id string) string {
	if name == "" {
		return id
	}
	if id == "" {
		return name
	}
	return fmt.Sprintf("%s (%s)", name, id)
}
