package scheduler

import (
	"time"

	"github.com/charmbracelet/log"
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/mahcks/blockbusterr/internal/global"
	"github.com/mahcks/blockbusterr/internal/helpers"
)

type Scheduler struct {
	gctx       global.Context
	scheduler  gocron.Scheduler
	movieJobID string
	showJobID  string
}

// Setup initializes a new scheduler instance
func Setup(gctx global.Context, helpers helpers.Helpers) *Scheduler {
	svc := &Scheduler{
		gctx: gctx,
	}
	var err error

	svc.scheduler, err = gocron.NewScheduler()
	if err != nil {
		log.Error("[scheduler] Failed to create new scheduler", "error", err)
		return nil
	}

	movieInterval, err := gctx.Crate().SQL.Queries().GetMovieInterval(gctx)
	if err != nil {
		log.Error("[scheduler] Failed to get movie interval from database", "error", err)
		return nil
	}

	if !movieInterval.Valid {
		log.Error("[scheduler] Movie interval is not set!")
		return nil
	}

	// Skip the movie interval if it's set to 0
	if movieInterval.Int32 != 0 {
		svc.StartMovieJob(int(movieInterval.Int32), func() {
			svc.MovieJobFunc(gctx, helpers)
		})
	}

	sonarrInterval, err := gctx.Crate().SQL.Queries().GetShowInterval(gctx)
	if err != nil {
		log.Error("[scheduler] Failed to get show interval from database", "error", err)
		return nil
	}

	if !sonarrInterval.Valid {
		log.Error("[scheduler] Show interval is not set!")
		return nil
	}

	// Skip the show interval if it's set to 0
	if sonarrInterval.Int32 != 0 {
		svc.StartShowJob(int(sonarrInterval.Int32), func() {
			svc.ShowJobFunc(gctx, helpers)
		})
	}

	return svc
}

// StartMovieJob starts a Movie job with a dynamic interval in hours
func (s *Scheduler) StartMovieJob(interval int, jobFunc func()) {
	// If a Movie job is already running, stop it first
	if s.movieJobID != "" {
		log.Info("[scheduler] Stopping existing Movie job...")
		s.StopJob(s.movieJobID)
	}

	// Schedule a new Movie job
	s.scheduleJob(interval, jobFunc, "movie")
}

// StartShowJob starts a Show job with a dynamic interval in hours
func (s *Scheduler) StartShowJob(interval int, jobFunc func()) {
	// If a Show job is already running, stop it first
	if s.showJobID != "" {
		log.Info("[scheduler] Stopping existing Show job...")
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
		log.Error("[scheduler] Failed to create new job", "error", err, "jobType", jobType)
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
	log.Info("[scheduler] Running", jobType, "job immediately")
	err = job.RunNow()
	if err != nil {
		log.Error("[scheduler] Failed to run", jobType, "job immediately", "error", err)
		return
	}

	log.Info("[scheduler] Job scheduled with interval (hours)", "jobType", jobType, "interval", interval)
}

// StopJob stops the job with the given job ID
func (s *Scheduler) StopJob(jobID string) {
	if jobID != "" {
		jobUUID, err := uuid.Parse(jobID)
		if err != nil {
			log.Error("[scheduler] Failed to parse job ID", "error", err)
			return
		}

		// Remove job by ID
		err = s.scheduler.RemoveJob(jobUUID)
		if err != nil {
			log.Error("[scheduler] Failed to remove job", "error", err)
			return
		}
		log.Info("[scheduler] Job stopped successfully")
	}
}

// UpdateMovieJobInterval allows dynamic interval changes for the Movie job
func (s *Scheduler) UpdateMovieJobInterval(newInterval int, jobFunc func()) {
	log.Info("[scheduler] Updating Movie job interval", "newInterval (hours)", newInterval)
	s.StartMovieJob(newInterval, jobFunc)
}

// UpdateShowJobInterval allows dynamic interval changes for the Show job
func (s *Scheduler) UpdateShowJobInterval(newInterval int, jobFunc func()) {
	log.Info("[scheduler] Updating Show job interval", "newInterval (hours)", newInterval)
	s.StartShowJob(newInterval, jobFunc)
}
