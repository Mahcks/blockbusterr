package aggregator

import (
	"sort"
	"time"

	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/database"
	"github.com/mahcks/blockbusterr/internal/scoring"
)

// Aggregator handles content aggregation, scoring, and limit tracking
type Aggregator struct {
	cfg *config.Config
	db  *database.Database
}

// JobMinimum represents a job's minimum pick requirement
type JobMinimum struct {
	JobName string
	Minimum int
}

// New creates a new Aggregator instance
func New(cfg *config.Config, db *database.Database) *Aggregator {
	return &Aggregator{
		cfg: cfg,
		db:  db,
	}
}

// ShouldUseGlobalLimits returns true if global limits should be enforced for the given media type
func (a *Aggregator) ShouldUseGlobalLimits(mediaType string) bool {
	if mediaType == "movie" {
		return a.cfg.Jobs.GlobalLimitMovies > 0
	}
	if mediaType == "show" {
		return a.cfg.Jobs.GlobalLimitShows > 0
	}
	return false
}

// RankAndLimit sorts items by score and applies global limits if configured
// Returns the top items within the limit (or all items if no limit)
func (a *Aggregator) RankAndLimit(items []scoring.ContentScore, mediaType string) []scoring.ContentScore {
	// Sort by score descending
	sort.Slice(items, func(i, j int) bool {
		return items[i].Score > items[j].Score
	})

	// Assign ranks
	for i := range items {
		items[i].Rank = i + 1
	}

	// Apply limit if configured
	var limit int
	switch mediaType {
	case "movie":
		limit = a.cfg.Jobs.GlobalLimitMovies
	case "show":
		limit = a.cfg.Jobs.GlobalLimitShows
	}

	// If no limit or limit is larger than items, return all
	if limit == 0 || limit >= len(items) {
		return items
	}

	// Return only top N items
	return items[:limit]
}

// ApplyMinimumPicks ensures each job gets at least its configured minimum picks
// Takes ALL ranked items (not just top N) and applies minimum picks logic
func (a *Aggregator) ApplyMinimumPicks(allItems []scoring.ContentScore, jobMinimums []JobMinimum, limit int) []scoring.ContentScore {
	if limit == 0 || len(jobMinimums) == 0 || len(allItems) == 0 {
		// No limit, no minimums, or no items
		return allItems
	}

	// Make a copy to avoid modifying original
	items := make([]scoring.ContentScore, len(allItems))
	copy(items, allItems)

	// Start with top N items
	topN := items
	if len(items) > limit {
		topN = items[:limit]
	}

	// Count current picks per job in top N
	jobCounts := make(map[string]int)
	for _, item := range topN {
		jobCounts[item.JobName]++
	}

	// Track which indices are in the final selection
	selected := make(map[int]bool)
	for i := 0; i < len(topN); i++ {
		selected[i] = true
	}

	// For each job with a minimum requirement
	for _, jobMin := range jobMinimums {
		if jobMin.Minimum == 0 {
			continue
		}

		currentCount := jobCounts[jobMin.JobName]
		if currentCount >= jobMin.Minimum {
			// Job already meets minimum
			continue
		}

		// Find more items from this job to meet minimum
		needed := jobMin.Minimum - currentCount
		added := 0

		// Look through ALL items (including those below the top N cutoff)
		for i := 0; i < len(items) && added < needed; i++ {
			if items[i].JobName == jobMin.JobName && !selected[i] {
				// Found an unselected item from this job
				// Find the lowest-scored currently selected item to replace (that's not from the same job)
				lowestIdx := -1
				lowestScore := 1.1 // Higher than max possible score
				for j := 0; j < len(items); j++ {
					if selected[j] && items[j].JobName != jobMin.JobName && items[j].Score < lowestScore {
						lowestIdx = j
						lowestScore = items[j].Score
					}
				}

				if lowestIdx != -1 {
					// Replace the lowest with this minimum pick
					selected[lowestIdx] = false
					selected[i] = true
					jobCounts[items[lowestIdx].JobName]--
					jobCounts[jobMin.JobName]++
					added++
				}
			}
		}
	}

	// Build result from selected items
	result := make([]scoring.ContentScore, 0, limit)
	for i := 0; i < len(items); i++ {
		if selected[i] {
			result = append(result, items[i])
		}
	}

	// Re-sort by score and re-rank
	sort.Slice(result, func(i, j int) bool {
		return result[i].Score > result[j].Score
	})

	for i := range result {
		result[i].Rank = i + 1
	}

	return result
}

// CalculateResetTime calculates the next reset time based on the period
func (a *Aggregator) CalculateResetTime(period string) time.Time {
	now := time.Now()

	switch period {
	case "sync":
		// Reset after each sync - effectively no carry-over
		return now
	case "daily":
		// Reset at midnight tomorrow
		return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	case "weekly":
		// Reset next Monday at midnight
		daysUntilMonday := (8 - int(now.Weekday())) % 7
		if daysUntilMonday == 0 {
			daysUntilMonday = 7
		}
		return time.Date(now.Year(), now.Month(), now.Day()+daysUntilMonday, 0, 0, 0, 0, now.Location())
	case "monthly":
		// Reset on the 1st of next month at midnight
		if now.Month() == 12 {
			return time.Date(now.Year()+1, 1, 1, 0, 0, 0, 0, now.Location())
		}
		return time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
	default:
		// Default to sync (immediate reset)
		return now
	}
}

// CheckGlobalLimit checks if adding an item would exceed the global limit for the period
// Returns true if the item can be added, false if limit is reached
func (a *Aggregator) CheckGlobalLimit(mediaType string) (bool, error) {
	if !a.ShouldUseGlobalLimits(mediaType) {
		return true, nil // No limit configured
	}

	period := a.cfg.Jobs.GlobalPeriod
	if period == "" {
		period = "sync"
	}

	currentCount, err := a.db.GetCurrentGlobalCount(period, mediaType)
	if err != nil {
		return false, err
	}

	var limit int
	if mediaType == "movie" {
		limit = a.cfg.Jobs.GlobalLimitMovies
	} else if mediaType == "show" {
		limit = a.cfg.Jobs.GlobalLimitShows
	}

	return currentCount < limit, nil
}

// IncrementGlobalCount increments the global count for the period and media type
func (a *Aggregator) IncrementGlobalCount(mediaType string) error {
	if !a.ShouldUseGlobalLimits(mediaType) {
		return nil // No limit configured, nothing to track
	}

	period := a.cfg.Jobs.GlobalPeriod
	if period == "" {
		period = "sync"
	}

	resetAt := a.CalculateResetTime(period)
	return a.db.IncrementGlobalLimit(period, mediaType, resetAt)
}
