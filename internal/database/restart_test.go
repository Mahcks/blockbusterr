package database

import (
	"github.com/mahcks/blockbusterr/pkg/enums"
	"testing"
	"time"
)

func TestRestartRecoversInterruptedHistoryWithoutForgettingDelivery(t *testing.T) {
	dir := t.TempDir()
	db, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.StartJobRun("job", "Job", "movie", "radarr", time.Now()); err != nil {
		t.Fatal(err)
	}
	if _, err = db.StartSelectionCycle(time.Now()); err != nil {
		t.Fatal(err)
	}
	done, err := db.StartSelectionCycle(time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err = db.CompleteSelectionCycle(done, enums.SelectionCycleCompleted, 1, 0, 1, 0, 0, ""); err != nil {
		t.Fatal(err)
	}
	if err = db.LogActivity(ActivityLog{Timestamp: time.Now(), MediaType: "movie", Title: "Delivered", TMDBID: 42, Status: string(enums.ActivityStatusAdded)}); err != nil {
		t.Fatal(err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = New(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	runs, err := db.GetRecentJobRuns(10, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(runs) != 1 || runs[0].Status != string(enums.JobRunStatusFailed) || runs[0].FinishedAt == nil || runs[0].ErrorMessage == "" {
		t.Fatalf("runs=%+v", runs)
	}
	cycles, err := db.GetRecentSelectionCycles(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(cycles) != 2 {
		t.Fatalf("cycles=%+v", cycles)
	}
	for _, cycle := range cycles {
		if cycle.ID == done {
			if cycle.Status != enums.SelectionCycleCompleted || !cycle.AccountingComplete {
				t.Fatalf("completed cycle changed: %+v", cycle)
			}
		} else if cycle.Status != enums.SelectionCycleFailed || cycle.FinishedAt == nil || cycle.AccountingComplete {
			t.Fatalf("interrupted cycle=%+v", cycle)
		}
	}
	if _, err = db.ClearActivityHistory(); err != nil {
		t.Fatal(err)
	}
	if _, found, err := db.LatestSuccessfulDelivery("movie", 42, 0); err != nil || !found {
		t.Fatalf("delivery memory found=%v err=%v", found, err)
	}
	runs, err = db.GetRecentJobRuns(10, "")
	if err != nil || len(runs) != 0 {
		t.Fatalf("recovered history not clearable: %+v %v", runs, err)
	}
}
