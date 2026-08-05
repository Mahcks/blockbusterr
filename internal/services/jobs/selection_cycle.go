package jobs

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/database"
	"github.com/mahcks/blockbusterr/internal/integrations"
	"github.com/mahcks/blockbusterr/pkg/enums"
)

type SelectionCyclePlan struct {
	Movies       SelectionAllocation    `json:"movies"`
	Shows        SelectionAllocation    `json:"shows"`
	Participants []SelectionParticipant `json:"participants,omitempty"`
	Duplicates   int                    `json:"duplicates_merged,omitempty"`
	Errors       map[string]string      `json:"errors,omitempty"`
}

type SelectionParticipant struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Source        string `json:"source"`
	MediaType     string `json:"media_type"`
	Found         int    `json:"found"`
	Candidates    int    `json:"candidates"`
	Rejected      int    `json:"rejected"`
	AlreadyExists int    `json:"already_exists"`
	RepeatBlocked int    `json:"repeat_blocked"`
	Capped        int    `json:"capped"`
	DeliveryLimit int    `json:"delivery_limit"`
}

type selectionExecution struct {
	scores map[string]ScoreInfo
	movies []integrations.Movie
	shows  []integrations.Show
}

type selectionJobRunner func(context.Context, *config.Config, *database.Database, config.DynamicJob, bool, map[string]ScoreInfo, []integrations.Movie, []integrations.Show) (JobExecutionSummary, error)

type selectionCycleOutcome struct {
	movieDelivered int
	showDelivered  int
	failedItems    int
}

func PreviewSelectionCycle(cfg *config.Config, db *database.Database) (SelectionCyclePlan, error) {
	return planSelectionCycle(cfg, db)
}

func runSelectionCycle(ctx context.Context, cfg *config.Config, db *database.Database, dryRun bool, run selectionJobRunner) (SelectionCyclePlan, error) {
	cycleID := int64(0)
	if db != nil {
		var err error
		cycleID, err = db.StartSelectionCycle(time.Now())
		if err != nil {
			return SelectionCyclePlan{}, err
		}
	}

	plan, err := planSelectionCycle(cfg, db)
	if err != nil {
		if db != nil {
			_ = db.CompleteSelectionCycle(cycleID, enums.SelectionCycleFailed, 0, 0, 0, 0, 0, err.Error())
		}
		return plan, err
	}
	if db != nil {
		if err := db.SaveSelectionCycleItems(cycleID, selectionCycleItems(plan)); err != nil {
			_ = db.CompleteSelectionCycle(cycleID, enums.SelectionCycleFailed, 0, 0, 0, 0, 0, err.Error())
			return plan, err
		}
	}
	outcome := executeSelectionPlan(ctx, cfg, db, dryRun, &plan, run)
	if db != nil {
		messages := make([]string, 0, len(plan.Errors))
		for jobID, message := range plan.Errors {
			messages = append(messages, jobID+": "+message)
		}
		slices.Sort(messages)
		planned := len(plan.Movies.Winners) + len(plan.Shows.Winners)
		status := selectionCycleStatus(planned, outcome.movieDelivered+outcome.showDelivered, outcome.failedItems, len(plan.Errors), ctx.Err())
		_ = db.CompleteSelectionCycle(cycleID, status, len(plan.Movies.Winners), len(plan.Shows.Winners), outcome.movieDelivered, outcome.showDelivered, outcome.failedItems, strings.Join(messages, "; "))
	}
	return plan, ctx.Err()
}

func executeSelectionPlan(ctx context.Context, cfg *config.Config, db *database.Database, dryRun bool, plan *SelectionCyclePlan, run selectionJobRunner) selectionCycleOutcome {
	jobsByID := map[string]config.DynamicJob{}
	for _, job := range cfg.Jobs.List {
		jobsByID[job.ID] = job
	}
	selected := map[string]*selectionExecution{}
	jobOrder := []string{}
	appendWinners := func(winners []SelectionResult) {
		for rank, winner := range winners {
			jobID := winner.Candidate.JobID
			if selected[jobID] == nil {
				selected[jobID] = &selectionExecution{scores: map[string]ScoreInfo{}}
				jobOrder = append(jobOrder, jobID)
			}
			execution := selected[jobID]
			execution.scores[winner.Candidate.Key] = ScoreInfo{Score: winner.Candidate.Score, Rank: rank + 1}
			if winner.Candidate.Movie != nil {
				execution.movies = append(execution.movies, *winner.Candidate.Movie)
			}
			if winner.Candidate.Show != nil {
				execution.shows = append(execution.shows, *winner.Candidate.Show)
			}
		}
	}
	appendWinners(plan.Movies.Winners)
	appendWinners(plan.Shows.Winners)
	outcome := selectionCycleOutcome{}
	for _, jobID := range jobOrder {
		execution := selected[jobID]
		job, ok := jobsByID[jobID]
		planned := len(execution.scores)
		if !ok || !job.Enabled || !job.SelectionCycle {
			plan.Errors[jobID] = "job was deleted, disabled, or removed from ranked selection after planning"
			outcome.failedItems += planned
			continue
		}
		if ctx.Err() != nil {
			plan.Errors[jobID] = ctx.Err().Error()
			outcome.failedItems += planned
			continue
		}
		if len(execution.movies)+len(execution.shows) != len(execution.scores) {
			plan.Errors[jobID] = "ranked selection winner snapshot is incomplete"
			outcome.failedItems += planned
			continue
		}
		summary, err := run(ctx, cfg, db, job, dryRun, execution.scores, execution.movies, execution.shows)
		if job.MediaType == "show" {
			outcome.showDelivered += summary.Delivered()
		} else {
			outcome.movieDelivered += summary.Delivered()
		}
		outcome.failedItems += summary.Failed
		if err != nil {
			plan.Errors[jobID] = err.Error()
			unaccounted := planned - summary.Delivered() - summary.Skipped - summary.Failed
			outcome.failedItems += max(unaccounted, 0)
		}
	}
	return outcome
}

func selectionCycleStatus(planned, delivered, failedItems, errorCount int, executionErr error) enums.SelectionCycleStatus {
	if executionErr != nil || (planned == 0 && errorCount > 0) || (planned > 0 && delivered == 0 && failedItems >= planned) {
		return enums.SelectionCycleFailed
	}
	return enums.SelectionCycleCompleted
}

func selectionCycleItems(plan SelectionCyclePlan) []database.SelectionCycleItem {
	items := []database.SelectionCycleItem{}
	appendResults := func(results []SelectionResult, winner bool) {
		for index, result := range results {
			rank := 0
			if winner {
				rank = index + 1
			}
			items = append(items, database.SelectionCycleItem{MediaKey: result.Candidate.Key, JobID: result.Candidate.JobID, JobIDs: result.JobIDs, Sources: result.Sources, Score: result.Candidate.Score, Rank: rank, Reason: result.Reason, Snapshot: result.Candidate})
		}
	}
	appendResults(plan.Movies.Winners, true)
	appendResults(plan.Movies.Excluded, false)
	appendResults(plan.Shows.Winners, true)
	appendResults(plan.Shows.Excluded, false)
	return items
}

func planSelectionCycle(cfg *config.Config, db *database.Database) (SelectionCyclePlan, error) {
	movieCapacity, err := remainingSelectionCapacity(cfg, db, "movie", cfg.Jobs.Selection.MovieLimit)
	if err != nil {
		return SelectionCyclePlan{}, err
	}
	showCapacity, err := remainingSelectionCapacity(cfg, db, "show", cfg.Jobs.Selection.ShowLimit)
	if err != nil {
		return SelectionCyclePlan{}, err
	}
	return planSelectionCycleWithPreview(cfg, func(job config.DynamicJob) (PreviewResponse, error) {
		return PreviewDynamicJob(cfg, db, job)
	}, movieCapacity, showCapacity)
}

func planSelectionCycleWithPreview(cfg *config.Config, previewJob func(config.DynamicJob) (PreviewResponse, error), movieCapacity, showCapacity int) (SelectionCyclePlan, error) {
	plan := SelectionCyclePlan{Errors: map[string]string{}}
	if !cfg.Jobs.Selection.Enabled {
		return plan, fmt.Errorf("ranked selection is disabled")
	}
	if !cfg.Scoring.Enabled {
		return plan, fmt.Errorf("ranked selection requires content scoring")
	}
	movieCandidates, showCandidates := []SelectionCandidate{}, []SelectionCandidate{}
	movieMinima, showMinima := map[string]int{}, map[string]int{}
	jobLimits := map[string]int{}
	for _, job := range cfg.Jobs.List {
		if !job.Enabled || !job.SelectionCycle {
			continue
		}
		participant := SelectionParticipant{ID: job.ID, Name: job.Name, Source: job.Source, MediaType: job.MediaType, DeliveryLimit: job.DeliveryLimit}
		plan.Participants = append(plan.Participants, participant)
		participantIndex := len(plan.Participants) - 1
		if !IsJobSourceConfigured(cfg, job) {
			plan.Errors[job.ID] = "discovery source is not configured"
			continue
		}
		if job.Type == "smart_popular" {
			plan.Errors[job.ID] = "adaptive smart jobs cannot join ranked selection"
			continue
		}
		preview, err := previewJob(job)
		if err != nil {
			plan.Errors[job.ID] = err.Error()
			continue
		}
		if job.MediaType == "show" {
			showMinima[job.ID] = job.MinimumPicks
		} else {
			movieMinima[job.ID] = job.MinimumPicks
		}
		jobLimits[job.ID] = job.DeliveryLimit
		for _, item := range preview.Items {
			plan.Participants[participantIndex].Found++
			if item.FilteredOut {
				plan.Participants[participantIndex].Rejected++
				continue
			}
			if item.AlreadyExists {
				if item.RepeatBlocked {
					plan.Participants[participantIndex].RepeatBlocked++
				} else {
					plan.Participants[participantIndex].AlreadyExists++
				}
				continue
			}
			candidate := SelectionCandidate{JobID: job.ID, Source: job.Source, Title: item.Title, Year: item.Year, Score: item.Score, ProviderRank: item.ProviderRank, FilterChecks: item.FilterChecks, DecisionReason: item.DecisionReason}
			if job.MediaType == "show" {
				candidate.Key = integrations.ShowKey(integrations.IDs{TVDB: item.TVDBID, TMDB: item.TMDBID, IMDB: item.IMDBID})
				candidate.Show = item.show
				if candidate.Key != "" {
					showCandidates = append(showCandidates, candidate)
					plan.Participants[participantIndex].Candidates++
				}
			} else {
				candidate.Key = integrations.MovieKey(integrations.IDs{TMDB: item.TMDBID, IMDB: item.IMDBID})
				candidate.Movie = item.movie
				if candidate.Key != "" {
					movieCandidates = append(movieCandidates, candidate)
					plan.Participants[participantIndex].Candidates++
				}
			}
		}
	}
	if len(plan.Participants) == 0 {
		return plan, fmt.Errorf("no enabled jobs are included in ranked selection")
	}
	for index := range plan.Participants {
		participant := &plan.Participants[index]
		if participant.DeliveryLimit > 0 && participant.Candidates > participant.DeliveryLimit {
			participant.Capped = participant.Candidates - participant.DeliveryLimit
		}
	}
	movieCandidates = limitSelectionCandidates(movieCandidates, jobLimits)
	showCandidates = limitSelectionCandidates(showCandidates, jobLimits)
	plan.Duplicates = duplicateSelectionCandidates(movieCandidates) + duplicateSelectionCandidates(showCandidates)
	var err error
	fullMovies, err := AllocateSelection(movieCandidates, cfg.Jobs.Selection.MovieLimit, movieMinima)
	if err != nil {
		return plan, err
	}
	fullShows, err := AllocateSelection(showCandidates, cfg.Jobs.Selection.ShowLimit, showMinima)
	if err != nil {
		return plan, err
	}
	plan.Movies, err = AllocateSelection(movieCandidates, movieCapacity, fitMinimums(movieMinima, movieCapacity))
	if err != nil {
		return plan, err
	}
	plan.Shows, err = AllocateSelection(showCandidates, showCapacity, fitMinimums(showMinima, showCapacity))
	if movieCapacity < cfg.Jobs.Selection.MovieLimit {
		markBudgetExclusions(plan.Movies.Excluded, fullMovies.Winners)
	}
	if showCapacity < cfg.Jobs.Selection.ShowLimit {
		markBudgetExclusions(plan.Shows.Excluded, fullShows.Winners)
	}
	return plan, err
}

func duplicateSelectionCandidates(candidates []SelectionCandidate) int {
	seen := map[string]bool{}
	for _, candidate := range candidates {
		seen[candidate.Key] = true
	}
	return len(candidates) - len(seen)
}

func remainingSelectionCapacity(cfg *config.Config, db *database.Database, mediaType string, cycleLimit int) (int, error) {
	globalLimit := cfg.Jobs.GlobalLimitMovies
	if mediaType == "show" {
		globalLimit = cfg.Jobs.GlobalLimitShows
	}
	if globalLimit <= 0 || db == nil {
		return cycleLimit, nil
	}
	used, err := db.CountDeliveriesSince(mediaType, time.Now().Add(-deliveryBudgetWindow(cfg.Jobs.GlobalPeriod)))
	if err != nil {
		return 0, err
	}
	return min(cycleLimit, max(0, globalLimit-used)), nil
}

func limitSelectionCandidates(candidates []SelectionCandidate, limits map[string]int) []SelectionCandidate {
	ordered := slices.Clone(candidates)
	slices.SortFunc(ordered, compareSelectionCandidates)
	used, result := map[string]int{}, make([]SelectionCandidate, 0, len(ordered))
	for _, candidate := range ordered {
		if limit := limits[candidate.JobID]; limit > 0 && used[candidate.JobID] >= limit {
			continue
		}
		used[candidate.JobID]++
		result = append(result, candidate)
	}
	return result
}

func fitMinimums(minima map[string]int, capacity int) map[string]int {
	result := map[string]int{}
	jobIDs := make([]string, 0, len(minima))
	for jobID := range minima {
		jobIDs = append(jobIDs, jobID)
	}
	slices.Sort(jobIDs)
	for remaining := true; remaining && capacity > 0; {
		remaining = false
		for _, jobID := range jobIDs {
			if result[jobID] >= minima[jobID] {
				continue
			}
			result[jobID]++
			capacity--
			remaining = capacity > 0
			if capacity == 0 {
				break
			}
		}
	}
	return result
}

func markBudgetExclusions(results, uncappedWinners []SelectionResult) {
	winnerKeys := map[string]bool{}
	for _, winner := range uncappedWinners {
		winnerKeys[winner.Candidate.Key] = true
	}
	for index := range results {
		if winnerKeys[results[index].Candidate.Key] {
			results[index].Reason = enums.SelectionReasonBudget
		}
	}
}
