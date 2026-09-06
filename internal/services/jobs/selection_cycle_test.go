package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/database"
	"github.com/mahcks/blockbusterr/internal/integrations"
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
	defer func() { _ = db.Close() }()
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
	coordinator := NewExecutionCoordinator(t.Context())
	release := make(chan struct{})
	if err := coordinator.Start(t.Context(), SelectionCycleExecutionID, func(context.Context) error {
		<-release
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := coordinator.RunSelectionCycle(t.Context(), &config.Config{}, nil, true); !errors.Is(err, ErrExecutionAlreadyRunning) {
		t.Fatal("expected concurrent cycle to be rejected")
	}
	close(release)
	coordinator.Stop()
}

func TestSelectionCandidateRoundTripPreservesShowIdentityWithoutTVDB(t *testing.T) {
	want := SelectionCandidate{Key: "show:tmdb:42", JobID: "shows", Source: "tmdb", Score: .9, Show: &integrations.Show{Title: "Exact Show", IDs: integrations.IDs{TMDB: 42, IMDB: "tt42"}}}
	encoded, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	var got SelectionCandidate
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatal(err)
	}
	if got.Key != want.Key || got.Show == nil || got.Show.IDs.TMDB != 42 || got.Show.IDs.TVDB != 0 || got.Show.IDs.IMDB != "tt42" {
		t.Fatalf("round trip candidate = %+v", got)
	}
}

func TestExecuteSelectionPlanUsesStoredWinnerAndTruthfulOutcome(t *testing.T) {
	movie := integrations.Movie{Title: "Planned", IDs: integrations.IDs{TMDB: 42}}
	plan := SelectionCyclePlan{Errors: map[string]string{}, Movies: SelectionAllocation{Winners: []SelectionResult{{Candidate: SelectionCandidate{Key: "movie:tmdb:42", JobID: "job", Movie: &movie, Score: .8}}}}}
	cfg := &config.Config{}
	cfg.Jobs.List = []config.DynamicJob{{ID: "job", Enabled: true, SelectionCycle: true, MediaType: "movie"}}
	calls := 0
	outcome := executeSelectionPlan(t.Context(), cfg, nil, false, &plan, func(_ context.Context, _ *config.Config, _ *database.Database, _ config.DynamicJob, _ bool, scores map[string]ScoreInfo, movies []integrations.Movie, shows []integrations.Show) (JobExecutionSummary, error) {
		calls++
		if len(movies) != 1 || movies[0].IDs.TMDB != 42 || len(shows) != 0 || scores["movie:tmdb:42"].Score != .8 {
			t.Fatalf("execution snapshot: scores=%v movies=%+v shows=%+v", scores, movies, shows)
		}
		return JobExecutionSummary{Added: 1}, nil
	})
	if calls != 1 || outcome.movieDelivered != 1 || outcome.failedItems != 0 || selectionCycleStatus(1, 1, 0, 0, nil) != enums.SelectionCycleCompleted {
		t.Fatalf("calls=%d outcome=%+v", calls, outcome)
	}

	disabled := plan
	disabled.Errors = map[string]string{}
	cfg.Jobs.List[0].Enabled = false
	outcome = executeSelectionPlan(t.Context(), cfg, nil, false, &disabled, func(context.Context, *config.Config, *database.Database, config.DynamicJob, bool, map[string]ScoreInfo, []integrations.Movie, []integrations.Show) (JobExecutionSummary, error) {
		t.Fatal("disabled job executed")
		return JobExecutionSummary{}, nil
	})
	if outcome.failedItems != 1 || disabled.Errors["job"] == "" || selectionCycleStatus(1, 0, 1, 1, nil) != enums.SelectionCycleFailed {
		t.Fatalf("disabled outcome=%+v errors=%v", outcome, disabled.Errors)
	}
}

func TestSelectionCycleStatusCoversPartialFailureAndCancellation(t *testing.T) {
	if got := selectionCycleStatus(2, 1, 1, 1, nil); got != enums.SelectionCycleCompleted {
		t.Fatalf("partial status = %s", got)
	}
	if got := selectionCycleStatus(2, 0, 2, 1, nil); got != enums.SelectionCycleFailed {
		t.Fatalf("all-failed status = %s", got)
	}
	if got := selectionCycleStatus(2, 1, 0, 0, context.Canceled); got != enums.SelectionCycleFailed {
		t.Fatalf("cancelled status = %s", got)
	}
}

func TestSelectionPlanDistinguishesUnlimitedAndExhaustedCapacity(t *testing.T) {
	for _, tc := range []struct {
		name                               string
		cycle, global, used, minimum, want int
	}{
		{"exhausted", 2, 1, 1, 1, 0},
		{"one slot", 2, 2, 1, 1, 1},
		{"unlimited", 0, 0, 0, 1, 2},
		{"unlimited cycle bounded globally", 0, 1, 0, 1, 1},
		{"unlimited cycle exhausted", 0, 1, 1, 1, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, err := database.New(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = db.Close() }()
			cfg := &config.Config{}
			cfg.Scoring.Enabled = true
			cfg.Jobs.Selection.Enabled = true
			cfg.Jobs.Selection.MovieLimit = tc.cycle
			cfg.Jobs.GlobalLimitMovies = tc.global
			cfg.Simkl.ClientID = "test"
			cfg.Jobs.List = []config.DynamicJob{{ID: "job", Enabled: true, SelectionCycle: true, Source: "simkl", MediaType: "movie", MinimumPicks: tc.minimum}}
			for i := 0; i < tc.used; i++ {
				if _, _, err := db.TryReserveDelivery(1, "previous", "movie", 0, tc.global, time.Now().Add(-24*time.Hour)); err != nil {
					t.Fatal(err)
				}
			}
			capacity, err := remainingSelectionCapacity(cfg, db, "movie", tc.cycle)
			if err != nil {
				t.Fatal(err)
			}
			plan, err := planSelectionCycleWithPreview(cfg, func(config.DynamicJob) (PreviewResponse, error) {
				return PreviewResponse{Items: []PreviewItem{{TMDBID: 1, Score: .8}, {TMDBID: 2, Score: .7}}}, nil
			}, capacity, 0)
			if err != nil {
				t.Fatal(err)
			}
			if len(plan.Movies.Winners) != tc.want {
				t.Fatalf("winners=%d want=%d", len(plan.Movies.Winners), tc.want)
			}
			if tc.global > 0 && tc.want < 2 {
				for _, excluded := range plan.Movies.Excluded {
					if excluded.Reason != enums.SelectionReasonBudget {
						t.Fatalf("wrong exclusion: %+v", excluded)
					}
				}
			}
		})
	}
}
