package scheduler

import (
	"fmt"
	"time"

	"github.com/charmbracelet/log"
	"github.com/mahcks/blockbusterr/internal/global"
	"github.com/mahcks/blockbusterr/internal/helpers"
	"github.com/mahcks/blockbusterr/internal/notifications"
	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	gctx          global.Context
	cron          *cron.Cron
	notifications *notifications.NotificationManager
	helpers       helpers.Helpers
	movieJobIDs   map[string]cron.EntryID
	showJobIDs    map[string]cron.EntryID
}

// Setup initializes a new scheduler instance
func Setup(gctx global.Context, helpers helpers.Helpers, notifications *notifications.NotificationManager) *Scheduler {
	svc := &Scheduler{
		gctx:          gctx,
		notifications: notifications,
		helpers:       helpers,
		cron:          cron.New(),
		movieJobIDs:   make(map[string]cron.EntryID),
		showJobIDs:    make(map[string]cron.EntryID),
	}

	// Setup individual cron jobs for each movie list
	movieSettings, err := gctx.Crate().SQL.Queries().GetMovieSettings(gctx)
	if err != nil {
		log.Error("[Scheduler] Failed to retrieve movie settings from the database. Please check your database connection and try again.", "error", err)
		return nil
	}

	// Schedule each list job with its cron expression
	if movieSettings.CronAnticipated.Valid {
		svc.scheduleMovieJob(movieSettings.CronAnticipated.String, svc.AnticipatedJobFunc, "movie-anticipated")
	}

	if movieSettings.CronBoxOffice.Valid {
		svc.scheduleMovieJob(movieSettings.CronBoxOffice.String, svc.BoxOfficeJobFunc, "movie-box_office")
	}

	if movieSettings.CronPopular.Valid {
		svc.scheduleMovieJob(movieSettings.CronPopular.String, svc.PopularJobFunc, "movie-popular")
	}

	if movieSettings.CronTrending.Valid {
		svc.scheduleMovieJob(movieSettings.CronTrending.String, svc.TrendingJobFunc, "movie-trending")
	}

	// Setup individual cron jobs for each show list
	showSettings, err := gctx.Crate().SQL.Queries().GetShowSettings(gctx)
	if err != nil {
		log.Error("[Scheduler] Failed to retrieve show settings from the database. Please check your database connection and try again.", "error", err)
		return nil
	}

	// Schedule each show list job with its cron expression
	if showSettings.CronJobAnticipated.Valid {
		svc.scheduleShowJob(showSettings.CronJobAnticipated.String, svc.AnticipatedShowJobFunc, "show-anticipated")
	}

	if showSettings.CronJobPopular.Valid {
		svc.scheduleShowJob(showSettings.CronJobPopular.String, svc.PopularShowJobFunc, "show-popular")
	}

	if showSettings.CronJobTrending.Valid {
		svc.scheduleShowJob(showSettings.CronJobTrending.String, svc.TrendingShowJobFunc, "show-trending")
	}

	// Start the scheduler
	svc.cron.Start()
	log.Info("[Scheduler] Scheduler started successfully.")

	return svc
}

// scheduleMovieJob schedules a movie list job using a cron expression
func (s *Scheduler) scheduleMovieJob(cronExpr string, jobFunc func(), listType string) {
	// If a job is already scheduled, stop it first
	if jobID, exists := s.movieJobIDs[listType]; exists {
		s.cron.Remove(jobID)
		log.Infof("[Scheduler] Existing %s movie job stopped.", listType)
	}

	// Schedule the new job
	jobID, err := s.cron.AddFunc(cronExpr, jobFunc)
	if err != nil {
		log.Error("[Scheduler] Could not schedule %s movie job. Please check cron expression %s and verify your settings.", listType, cronExpr, "error", err)
		return
	}

	s.movieJobIDs[listType] = jobID
	log.Infof("[Scheduler] Successfully scheduled %s movie job with cron expression: %s.", listType, cronExpr)

	// Run the job immediately once after scheduling
	s.RunJobOnDemand(listType, true)
}

// scheduleShowJob schedules a show list job using a cron expression
func (s *Scheduler) scheduleShowJob(cronExpr string, jobFunc func(), listType string) {
	// If a job is already scheduled, stop it first
	if jobID, exists := s.showJobIDs[listType]; exists {
		s.cron.Remove(jobID)
		log.Infof("[Scheduler] Existing %s show job stopped.", listType)
	}

	// Schedule the new job
	jobID, err := s.cron.AddFunc(cronExpr, jobFunc)
	if err != nil {
		log.Error("[Scheduler] Could not schedule %s show job. Please check cron expression %s and verify your settings.", listType, cronExpr, "error", err)
		return
	}

	s.showJobIDs[listType] = jobID
	log.Infof("[Scheduler] Successfully scheduled %s show job with cron expression: %s.", listType, cronExpr)

	// Run the job immediately once after scheduling
	s.RunJobOnDemand(listType, false)
}

// StopJob stops a specific movie job by listType
func (s *Scheduler) StopJob(listType string, isMovie bool) {
	if isMovie {
		if jobID, exists := s.movieJobIDs[listType]; exists {
			s.cron.Remove(jobID)
			log.Infof("[Scheduler] Successfully stopped %s movie job.", listType)
		} else {
			log.Warnf("[Scheduler] No %s movie job found to stop.", listType)
		}
	} else {
		if jobID, exists := s.showJobIDs[listType]; exists {
			s.cron.Remove(jobID)
			log.Infof("[Scheduler] Successfully stopped %s show job.", listType)
		} else {
			log.Warnf("[Scheduler] No %s show job found to stop.", listType)
		}
	}
}

// JobStatus holds information about the current state of a job
type JobStatus struct {
	JobID   string    `json:"job_id"`
	JobType string    `json:"job_type"`
	LastRun time.Time `json:"last_run"`
	NextRun time.Time `json:"next_run"`
}

// GetJobStatus returns the status of the movie and show jobs
func (s *Scheduler) GetJobStatus() []JobStatus {
	var statuses []JobStatus

	// Movie Job Statuses
	for listType, jobID := range s.movieJobIDs {
		entry := s.cron.Entry(jobID)
		statuses = append(statuses, JobStatus{
			JobID:   fmt.Sprintf("%d", jobID),
			JobType: listType,
			LastRun: entry.Prev,
			NextRun: entry.Next,
		})
	}

	// Show Job Statuses
	for listType, jobID := range s.showJobIDs {
		entry := s.cron.Entry(jobID)
		statuses = append(statuses, JobStatus{
			JobID:   fmt.Sprintf("%d", jobID),
			JobType: listType,
			LastRun: entry.Prev,
			NextRun: entry.Next,
		})
	}

	return statuses
}

// RunJobOnDemand runs a specific job immediately without affecting the cron schedule
func (s *Scheduler) RunJobOnDemand(listType string, isMovie bool) error {
	if isMovie {
		if jobID, exists := s.movieJobIDs[listType]; exists {
			log.Infof("[Scheduler] Manually triggered %s movie job.", listType)
			s.cron.Entry(jobID).Job.Run()
			return nil
		}
		return fmt.Errorf("[Scheduler] No movie job found for %s", listType)
	} else {
		if jobID, exists := s.showJobIDs[listType]; exists {
			log.Infof("[Scheduler] Manually triggered %s show job.", listType)
			s.cron.Entry(jobID).Job.Run()
			return nil
		}
		return fmt.Errorf("[Scheduler] No show job found for %s", listType)
	}
}
