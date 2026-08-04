package jobs

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/database"
	"github.com/mahcks/blockbusterr/pkg/enums"
)

func TestPlanSelectionCycleKeepsSuccessfulProviders(t *testing.T) {
	cfg := &config.Config{}
	cfg.Scoring.Enabled = true
	cfg.Jobs.Selection.Enabled = true
	cfg.Jobs.Selection.MovieLimit = 1
	cfg.Simkl.ClientID = "configured"
	cfg.Jobs.List = []config.DynamicJob{
		{ID: "good", Enabled: true, SelectionCycle: true, Source: "simkl", MediaType: "movie"},
		{ID: "failed", Enabled: true, SelectionCycle: true, Source: "simkl", MediaType: "movie"},
	}

	plan, err := planSelectionCycleWithPreview(cfg, func(job config.DynamicJob) (PreviewResponse, error) {
		if job.ID == "failed" {
			return PreviewResponse{}, errors.New("provider unavailable")
		}
		return PreviewResponse{Items: []PreviewItem{{TMDBID: 1, Score: .8, Rank: 1}}}, nil
	}, 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Movies.Winners) != 1 || plan.Movies.Winners[0].Candidate.JobID != "good" || plan.Errors["failed"] == "" || len(plan.Participants) != 2 || plan.Participants[0].Candidates != 1 {
		t.Fatalf("partial plan = %+v", plan)
	}
}

func TestPlanSelectionCycleRequiresParticipatingJob(t *testing.T) {
	cfg := &config.Config{}
	cfg.Scoring.Enabled = true
	cfg.Jobs.Selection.Enabled = true
	_, err := planSelectionCycleWithPreview(cfg, func(config.DynamicJob) (PreviewResponse, error) {
		t.Fatal("preview should not run without participating jobs")
		return PreviewResponse{}, nil
	}, 1, 1)
	if err == nil || err.Error() != "no enabled jobs are included in ranked selection" {
		t.Fatalf("err = %v", err)
	}
}

func TestSelectionCapacityHonorsRollingBudget(t *testing.T) {
	db, err := database.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	cfg := &config.Config{}
	cfg.Jobs.GlobalLimitMovies = 1
	cfg.Jobs.GlobalPeriod = "daily"
	if _, _, err := db.TryReserveDelivery(1, "job", "movie", 0, 1, time.Now().Add(-24*time.Hour)); err != nil {
		t.Fatal(err)
	}
	capacity, err := remainingSelectionCapacity(cfg, db, "movie", 5)
	if err != nil || capacity != 0 {
		t.Fatalf("capacity=%d err=%v", capacity, err)
	}
	if got := fitMinimums(map[string]int{"b": 2, "a": 2}, 3); !reflect.DeepEqual(got, map[string]int{"a": 2, "b": 1}) {
		t.Fatalf("minimum allocation=%v", got)
	}
	candidates := limitSelectionCandidates([]SelectionCandidate{{Key: "low", JobID: "job", Score: .2}, {Key: "high", JobID: "job", Score: .9}}, map[string]int{"job": 1})
	if len(candidates) != 1 || candidates[0].Key != "high" {
		t.Fatalf("per-job limit candidates=%+v", candidates)
	}
	excluded := []SelectionResult{{Candidate: SelectionCandidate{Key: "budget"}}, {Candidate: SelectionCandidate{Key: "displaced"}}}
	markBudgetExclusions(excluded, []SelectionResult{{Candidate: SelectionCandidate{Key: "budget"}}})
	if excluded[0].Reason != enums.SelectionReasonBudget || excluded[1].Reason == enums.SelectionReasonBudget {
		t.Fatalf("budget reasons=%+v", excluded)
	}
}

func TestRunSelectionCycleRejectsConcurrentRun(t *testing.T) {
	selectionCycleMutex.Lock()
	defer selectionCycleMutex.Unlock()
	if _, err := RunSelectionCycle(t.Context(), &config.Config{}, nil, true); err == nil {
		t.Fatal("expected concurrent cycle to be rejected")
	}
}
