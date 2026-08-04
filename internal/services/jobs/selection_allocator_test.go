package jobs

import (
	"reflect"
	"testing"

	"github.com/mahcks/blockbusterr/pkg/enums"
)

func TestAllocateSelection(t *testing.T) {
	candidates := []SelectionCandidate{
		{Key: "movie:1", JobID: "job-b", Source: "tmdb", Score: .9, ProviderRank: 2},
		{Key: "movie:1", JobID: "job-a", Source: "trakt", Score: .9, ProviderRank: 1},
		{Key: "movie:2", JobID: "job-a", Source: "trakt", Score: .8, ProviderRank: 2},
		{Key: "movie:3", JobID: "job-b", Source: "tmdb", Score: .7, ProviderRank: 1},
	}

	allocation, err := AllocateSelection(candidates, 2, map[string]int{"job-b": 1})
	if err != nil {
		t.Fatal(err)
	}
	if got := []string{allocation.Winners[0].Candidate.Key, allocation.Winners[1].Candidate.Key}; !reflect.DeepEqual(got, []string{"movie:1", "movie:2"}) {
		t.Fatalf("winners = %v", got)
	}
	if allocation.Winners[0].Reason != enums.SelectionReasonMinimum || allocation.Winners[0].Candidate.JobID != "job-b" {
		t.Fatalf("minimum winner = %+v", allocation.Winners[0])
	}
	if !reflect.DeepEqual(allocation.Winners[0].JobIDs, []string{"job-a", "job-b"}) || !reflect.DeepEqual(allocation.Winners[0].Sources, []string{"tmdb", "trakt"}) {
		t.Fatalf("merged provenance = %+v", allocation.Winners[0])
	}
	if len(allocation.Excluded) != 1 || allocation.Excluded[0].Candidate.Key != "movie:3" || allocation.Excluded[0].Reason != enums.SelectionReasonDisplaced {
		t.Fatalf("excluded = %+v", allocation.Excluded)
	}
}

func TestAllocateSelectionIsDeterministicAndValidatesMinima(t *testing.T) {
	candidates := []SelectionCandidate{
		{Key: "movie:b", JobID: "job", Score: .5},
		{Key: "movie:a", JobID: "job", Score: .5},
	}
	first, err := AllocateSelection(candidates, 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	second, err := AllocateSelection([]SelectionCandidate{candidates[1], candidates[0]}, 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	if first.Winners[0].Candidate.Key != "movie:a" || !reflect.DeepEqual(first, second) {
		t.Fatalf("non-deterministic allocation: %+v %+v", first, second)
	}
	if _, err := AllocateSelection(candidates, 1, map[string]int{"job": 2}); err == nil {
		t.Fatal("expected minima over capacity to fail")
	}
	empty, err := AllocateSelection(nil, 0, nil)
	if err != nil || len(empty.Winners) != 0 {
		t.Fatalf("empty allocation = %+v, %v", empty, err)
	}
	if _, err := AllocateSelection([]SelectionCandidate{{Key: "bad", JobID: "job", Score: 2}}, 1, nil); err == nil {
		t.Fatal("expected invalid normalized score to fail")
	}
}
