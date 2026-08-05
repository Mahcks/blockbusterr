package jobs

import (
	"cmp"
	"fmt"
	"math"
	"slices"

	"github.com/mahcks/blockbusterr/pkg/enums"
)

// SelectionCandidate is one rule-approved candidate from one job.
type SelectionCandidate struct {
	Key          string  `json:"key"`
	JobID        string  `json:"job_id"`
	Source       string  `json:"source"`
	Title        string  `json:"title"`
	Year         int     `json:"year,omitempty"`
	Score        float64 `json:"score"`
	ProviderRank int     `json:"provider_rank"`
}

type SelectionResult struct {
	Candidate SelectionCandidate    `json:"candidate"`
	Reason    enums.SelectionReason `json:"reason"`
	JobIDs    []string              `json:"job_ids"`
	Sources   []string              `json:"sources"`
}

type SelectionAllocation struct {
	Winners  []SelectionResult `json:"winners"`
	Excluded []SelectionResult `json:"excluded"`
}

// AllocateSelection deterministically deduplicates and selects candidates.
func AllocateSelection(candidates []SelectionCandidate, capacity int, minimumPicks map[string]int) (SelectionAllocation, error) {
	if capacity < 0 {
		return SelectionAllocation{}, fmt.Errorf("selection capacity cannot be negative")
	}
	if capacity == 0 {
		capacity = len(candidates)
	}
	minimumTotal := 0
	for jobID, minimum := range minimumPicks {
		if jobID == "" || minimum < 0 {
			return SelectionAllocation{}, fmt.Errorf("invalid minimum picks for job %q", jobID)
		}
		minimumTotal += minimum
	}
	if minimumTotal > capacity {
		return SelectionAllocation{}, fmt.Errorf("minimum picks total %d exceeds cycle capacity %d", minimumTotal, capacity)
	}
	for _, candidate := range candidates {
		if candidate.Key == "" || candidate.JobID == "" || math.IsNaN(candidate.Score) || candidate.Score < 0 || candidate.Score > 1 {
			return SelectionAllocation{}, fmt.Errorf("invalid selection candidate for job %q", candidate.JobID)
		}
	}

	ordered := slices.Clone(candidates)
	slices.SortFunc(ordered, compareSelectionCandidates)
	provenance := selectionProvenance(ordered)
	selected := map[string]bool{}
	allocation := SelectionAllocation{}

	jobIDs := make([]string, 0, len(minimumPicks))
	for jobID := range minimumPicks {
		jobIDs = append(jobIDs, jobID)
	}
	slices.Sort(jobIDs)
	for _, jobID := range jobIDs {
		remaining := minimumPicks[jobID]
		for _, candidate := range ordered {
			if remaining == 0 {
				break
			}
			if candidate.JobID != jobID || selected[candidate.Key] {
				continue
			}
			selected[candidate.Key] = true
			allocation.Winners = append(allocation.Winners, selectionResult(candidate, enums.SelectionReasonMinimum, provenance[candidate.Key]))
			remaining--
		}
	}

	for _, candidate := range ordered {
		if len(allocation.Winners) >= capacity {
			break
		}
		if selected[candidate.Key] {
			continue
		}
		selected[candidate.Key] = true
		allocation.Winners = append(allocation.Winners, selectionResult(candidate, enums.SelectionReasonWinner, provenance[candidate.Key]))
	}

	seenExcluded := map[string]bool{}
	for _, candidate := range ordered {
		if selected[candidate.Key] || seenExcluded[candidate.Key] {
			continue
		}
		seenExcluded[candidate.Key] = true
		allocation.Excluded = append(allocation.Excluded, selectionResult(candidate, enums.SelectionReasonDisplaced, provenance[candidate.Key]))
	}
	return allocation, nil
}

func compareSelectionCandidates(a, b SelectionCandidate) int {
	if score := cmp.Compare(b.Score, a.Score); score != 0 {
		return score
	}
	aRank, bRank := a.ProviderRank, b.ProviderRank
	if aRank == 0 {
		aRank = int(^uint(0) >> 1)
	}
	if bRank == 0 {
		bRank = int(^uint(0) >> 1)
	}
	if rank := cmp.Compare(aRank, bRank); rank != 0 {
		return rank
	}
	if key := cmp.Compare(a.Key, b.Key); key != 0 {
		return key
	}
	return cmp.Compare(a.JobID, b.JobID)
}

type candidateProvenance struct {
	jobs    []string
	sources []string
}

func selectionProvenance(candidates []SelectionCandidate) map[string]candidateProvenance {
	result := map[string]candidateProvenance{}
	for _, candidate := range candidates {
		item := result[candidate.Key]
		if !slices.Contains(item.jobs, candidate.JobID) {
			item.jobs = append(item.jobs, candidate.JobID)
		}
		if !slices.Contains(item.sources, candidate.Source) {
			item.sources = append(item.sources, candidate.Source)
		}
		result[candidate.Key] = item
	}
	return result
}

func selectionResult(candidate SelectionCandidate, reason enums.SelectionReason, provenance candidateProvenance) SelectionResult {
	slices.Sort(provenance.jobs)
	slices.Sort(provenance.sources)
	return SelectionResult{Candidate: candidate, Reason: reason, JobIDs: provenance.jobs, Sources: provenance.sources}
}
