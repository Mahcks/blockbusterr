package scheduler

import (
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
		log.Error("[scheduler] Failed to get movie settings from database", "error", err)
		return nil
	}

	// Schedule each list job with its cron expression
	if movieSettings.CronAnticipated.Valid {
		svc.scheduleMovieJob(movieSettings.CronAnticipated.String, svc.AnticipatedJobFunc, "Anticipated")
	}

	if movieSettings.CronBoxOffice.Valid {
		svc.scheduleMovieJob(movieSettings.CronBoxOffice.String, svc.BoxOfficeJobFunc, "Box Office")
	}

	if movieSettings.CronPopular.Valid {
		svc.scheduleMovieJob(movieSettings.CronPopular.String, svc.PopularJobFunc, "Popular")
	}

	if movieSettings.CronTrending.Valid {
		svc.scheduleMovieJob(movieSettings.CronTrending.String, svc.TrendingJobFunc, "Trending")
	}

	// Setup individual cron jobs for each show list
	showSettings, err := gctx.Crate().SQL.Queries().GetShowSettings(gctx)
	if err != nil {
		log.Error("[scheduler] Failed to get show settings from database", "error", err)
		return nil
	}

	// Schedule each show list job with its cron expression
	if showSettings.CronJobAnticipated.Valid {
		svc.scheduleShowJob(showSettings.CronJobAnticipated.String, svc.AnticipatedShowJobFunc, "Show Anticipated")
	}

	if showSettings.CronJobPopular.Valid {
		svc.scheduleShowJob(showSettings.CronJobPopular.String, svc.PopularShowJobFunc, "Show Popular")
	}

	if showSettings.CronJobTrending.Valid {
		svc.scheduleShowJob(showSettings.CronJobTrending.String, svc.TrendingShowJobFunc, "Show Trending")
	}

	// Start the scheduler
	svc.cron.Start()

	return svc
}

// scheduleMovieJob schedules a movie list job using a cron expression
func (s *Scheduler) scheduleMovieJob(cronExpr string, jobFunc func(), listType string) {
	// If a job is already scheduled, stop it first
	if jobID, exists := s.movieJobIDs[listType]; exists {
		s.cron.Remove(jobID)
	}

	// Schedule the new job
	jobID, err := s.cron.AddFunc(cronExpr, jobFunc)
	if err != nil {
		log.Error("[scheduler] Failed to schedule movie job", "listType", listType, "cronExpr", cronExpr, "error", err)
		return
	}

	s.movieJobIDs[listType] = jobID
	log.Infof("[scheduler] %s movie job scheduled with cron expression: %s", listType, cronExpr)

	// Run the job immediately once after scheduling
	log.Infof("[scheduler] Running %s movie job immediately", listType)
	jobFunc()
}

// scheduleShowJob schedules a show list job using a cron expression
func (s *Scheduler) scheduleShowJob(cronExpr string, jobFunc func(), listType string) {
	// If a job is already scheduled, stop it first
	if jobID, exists := s.showJobIDs[listType]; exists {
		s.cron.Remove(jobID)
	}

	// Schedule the new job
	jobID, err := s.cron.AddFunc(cronExpr, jobFunc)
	if err != nil {
		log.Error("[scheduler] Failed to schedule show job", "listType", listType, "cronExpr", cronExpr, "error", err)
		return
	}

	s.showJobIDs[listType] = jobID
	log.Infof("[scheduler] %s show job scheduled with cron expression: %s", listType, cronExpr)

	// Run the job immediately once after scheduling
	log.Infof("[scheduler] Running %s show job immediately", listType)
	jobFunc()
}

// StopJob stops a specific movie job by listType
func (s *Scheduler) StopJob(listType string, isMovie bool) {
	if isMovie {
		if jobID, exists := s.movieJobIDs[listType]; exists {
			s.cron.Remove(jobID)
			log.Infof("[scheduler] %s movie job stopped successfully", listType)
		}
	} else {
		if jobID, exists := s.showJobIDs[listType]; exists {
			s.cron.Remove(jobID)
			log.Infof("[scheduler] %s show job stopped successfully", listType)
		}
	}
}

// JobStatus holds information about the current state of a job
type JobStatus struct {
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
			JobType: listType,
			LastRun: entry.Prev,
			NextRun: entry.Next,
		})
	}

	// Show Job Statuses
	for listType, jobID := range s.showJobIDs {
		entry := s.cron.Entry(jobID)
		statuses = append(statuses, JobStatus{
			JobType: listType,
			LastRun: entry.Prev,
			NextRun: entry.Next,
		})
	}

	return statuses
}
