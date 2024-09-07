package scheduler

import (
	"log/slog"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/mahcks/blockbusterr/internal/global"
)

type Scheduler struct {
	gctx       global.Context
	scheduler  gocron.Scheduler
	movieJobID string
	showJobID  string
}

// Setup initializes a new scheduler instance
func Setup(gctx global.Context) *Scheduler {
	svc := &Scheduler{
		gctx: gctx,
	}
	var err error

	svc.scheduler, err = gocron.NewScheduler()
	if err != nil {
		slog.Error("[scheduler] Failed to create new scheduler", "error", err)
		return nil
	}

	// TODO: Initialize jobs here via database flags

	return svc
}

// StartMovieJob starts a Movie job with a dynamic interval in hours
func (s *Scheduler) StartMovieJob(interval int, jobFunc func()) {
	// If a Movie job is already running, stop it first
	if s.movieJobID != "" {
		slog.Info("[scheduler] Stopping existing Movie job...")
		s.StopJob(s.movieJobID)
	}

	// Schedule a new Movie job
	s.scheduleJob(interval, jobFunc, "movie")
}

// StartShowJob starts a Show job with a dynamic interval in hours
func (s *Scheduler) StartShowJob(interval int, jobFunc func()) {
	// If a Show job is already running, stop it first
	if s.showJobID != "" {
		slog.Info("[scheduler] Stopping existing Show job...")
		s.StopJob(s.showJobID)
	}

	// Schedule a new Show job
	s.scheduleJob(interval, jobFunc, "show")
}

// scheduleJob handles creating and starting a new job
func (s *Scheduler) scheduleJob(interval int, jobFunc func(), jobType string) {
	// Create the new job definition
	jobDefinition := gocron.DurationJob(time.Duration(interval) * time.Hour)
	task := gocron.NewTask(jobFunc)

	// Add the job to the scheduler
	job, err := s.scheduler.NewJob(jobDefinition, task)
	if err != nil {
		slog.Error("[scheduler] Failed to create new job", "error", err, "jobType", jobType)
		return
	}

	if jobType == "movie" {
		s.movieJobID = job.ID().String()
	} else if jobType == "show" {
		s.showJobID = job.ID().String()
	}

	// Start the scheduler asynchronously (non-blocking)
	s.scheduler.Start()

	// Run the job immediately once
	slog.Info("[scheduler] Running", jobType, "job immediately")
	err = job.RunNow()
	if err != nil {
		slog.Error("[scheduler] Failed to run", jobType, "job immediately", "error", err)
		return
	}

	slog.Info("[scheduler] Job scheduled with interval (hours)", "jobType", jobType, "interval", interval)
}

// StopJob stops the job with the given job ID
func (s *Scheduler) StopJob(jobID string) {
	if jobID != "" {
		jobUUID, err := uuid.Parse(jobID)
		if err != nil {
			slog.Error("[scheduler] Failed to parse job ID", "error", err)
			return
		}

		// Remove job by ID
		err = s.scheduler.RemoveJob(jobUUID)
		if err != nil {
			slog.Error("[scheduler] Failed to remove job", "error", err)
			return
		}
		slog.Info("[scheduler] Job stopped successfully")
	}
}

// UpdateMovieJobInterval allows dynamic interval changes for the Movie job
func (s *Scheduler) UpdateMovieJobInterval(newInterval int, jobFunc func()) {
	slog.Info("[scheduler] Updating Movie job interval", "newInterval (hours)", newInterval)
	s.StartMovieJob(newInterval, jobFunc)
}

// UpdateShowJobInterval allows dynamic interval changes for the Show job
func (s *Scheduler) UpdateShowJobInterval(newInterval int, jobFunc func()) {
	slog.Info("[scheduler] Updating Show job interval", "newInterval (hours)", newInterval)
	s.StartShowJob(newInterval, jobFunc)
}
