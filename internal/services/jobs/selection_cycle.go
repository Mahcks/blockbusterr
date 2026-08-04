package jobs

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/database"
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

var selectionCycleMutex sync.Mutex

func PreviewSelectionCycle(cfg *config.Config, db *database.Database) (SelectionCyclePlan, error) {
	return planSelectionCycle(cfg, db)
}

func RunSelectionCycle(ctx context.Context, cfg *config.Config, db *database.Database, dryRun bool) (SelectionCyclePlan, error) {
	if !selectionCycleMutex.TryLock() {
		return SelectionCyclePlan{}, fmt.Errorf("ranked selection cycle is already running")
	}
	defer selectionCycleMutex.Unlock()
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
			_ = db.CompleteSelectionCycle(cycleID, enums.SelectionCycleFailed, 0, 0, err.Error())
		}
		return plan, err
	}
	if db != nil {
		if err := db.SaveSelectionCycleItems(cycleID, selectionCycleItems(plan)); err != nil {
			_ = db.CompleteSelectionCycle(cycleID, enums.SelectionCycleFailed, 0, 0, err.Error())
			return plan, err
		}
	}
	jobsByID := map[string]config.DynamicJob{}
	for _, job := range cfg.Jobs.List {
		jobsByID[job.ID] = job
	}
	selected := map[string]map[string]ScoreInfo{}
	jobOrder := []string{}
	appendWinners := func(winners []SelectionResult) {
		for rank, winner := range winners {
			jobID := winner.Candidate.JobID
			if selected[jobID] == nil {
				selected[jobID] = map[string]ScoreInfo{}
				jobOrder = append(jobOrder, jobID)
			}
			selected[jobID][winner.Candidate.Key] = ScoreInfo{Score: winner.Candidate.Score, Rank: rank + 1}
		}
	}
	appendWinners(plan.Movies.Winners)
	appendWinners(plan.Shows.Winners)
	for _, jobID := range jobOrder {
		winners := selected[jobID]
		job, ok := jobsByID[jobID]
		if !ok || ctx.Err() != nil {
			continue
		}
		if err := RunSelectedDynamicJob(ctx, cfg, db, job, dryRun, winners); err != nil {
			plan.Errors[jobID] = err.Error()
		}
	}
	if db != nil {
		messages := make([]string, 0, len(plan.Errors))
		for jobID, message := range plan.Errors {
			messages = append(messages, jobID+": "+message)
		}
		slices.Sort(messages)
		status := enums.SelectionCycleCompleted
		if len(selected) == 0 && len(plan.Errors) > 0 {
			status = enums.SelectionCycleFailed
		}
		_ = db.CompleteSelectionCycle(cycleID, status, len(plan.Movies.Winners), len(plan.Shows.Winners), strings.Join(messages, "; "))
	}
	return plan, ctx.Err()
}

func selectionCycleItems(plan SelectionCyclePlan) []database.SelectionCycleItem {
	items := []database.SelectionCycleItem{}
	appendResults := func(results []SelectionResult, winner bool) {
		for index, result := range results {
			rank := 0
			if winner {
				rank = index + 1
			}
			items = append(items, database.SelectionCycleItem{MediaKey: result.Candidate.Key, JobID: result.Candidate.JobID, JobIDs: result.JobIDs, Sources: result.Sources, Score: result.Candidate.Score, Rank: rank, Reason: result.Reason})
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
			candidate := SelectionCandidate{JobID: job.ID, Source: job.Source, Title: item.Title, Year: item.Year, Score: item.Score, ProviderRank: item.ProviderRank}
			if job.MediaType == "show" {
				candidate.Key = showSelectionKey(item.TVDBID, item.TMDBID, item.IMDBID)
				if candidate.Key != "" {
					showCandidates = append(showCandidates, candidate)
					plan.Participants[participantIndex].Candidates++
				}
			} else {
				candidate.Key = movieSelectionKey(item.TMDBID, item.IMDBID)
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

func movieSelectionKey(tmdbID int, imdbID string) string {
	if tmdbID > 0 {
		return "movie:tmdb:" + strconv.Itoa(tmdbID)
	}
	if imdbID != "" {
		return "movie:imdb:" + imdbID
	}
	return ""
}

func showSelectionKey(tvdbID, tmdbID int, imdbID string) string {
	if tmdbID > 0 {
		return "show:tmdb:" + strconv.Itoa(tmdbID)
	}
	if tvdbID > 0 {
		return "show:tvdb:" + strconv.Itoa(tvdbID)
	}
	if imdbID != "" {
		return "show:imdb:" + imdbID
	}
	return ""
}
