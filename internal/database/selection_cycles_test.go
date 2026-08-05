package database

import (
	"testing"
	"time"

	"github.com/mahcks/blockbusterr/pkg/enums"
)

func TestSelectionCyclePersistence(t *testing.T) {
	db, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	cycleID, err := db.StartSelectionCycle(time.Now())
	if err != nil {
		t.Fatal(err)
	}
	items := []SelectionCycleItem{{MediaKey: "movie:tmdb:1", JobID: "job-a", JobIDs: []string{"job-a", "job-b"}, Sources: []string{"tmdb", "trakt"}, Score: .8, Rank: 1, Reason: enums.SelectionReasonWinner, Snapshot: map[string]any{"tmdb_id": 1}}}
	if err := db.SaveSelectionCycleItems(cycleID, items); err != nil {
		t.Fatal(err)
	}
	if err := db.CompleteSelectionCycle(cycleID, enums.SelectionCycleCompleted, 1, 0, 1, 0, 0, ""); err != nil {
		t.Fatal(err)
	}
	var status string
	var itemCount int
	var snapshot string
	if err := db.db.QueryRow("SELECT status FROM selection_cycles WHERE id = ?", cycleID).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if err := db.db.QueryRow("SELECT COUNT(*) FROM selection_cycle_items WHERE cycle_id = ?", cycleID).Scan(&itemCount); err != nil {
		t.Fatal(err)
	}
	if err := db.db.QueryRow("SELECT snapshot FROM selection_cycle_items WHERE cycle_id = ?", cycleID).Scan(&snapshot); err != nil {
		t.Fatal(err)
	}
	if status != string(enums.SelectionCycleCompleted) || itemCount != 1 || snapshot != `{"tmdb_id":1}` {
		t.Fatalf("status=%s items=%d snapshot=%s", status, itemCount, snapshot)
	}
	cycles, err := db.GetRecentSelectionCycles(1)
	if err != nil || len(cycles) != 1 || cycles[0].MovieWinners != 1 || cycles[0].MovieDelivered != 1 || !cycles[0].AccountingComplete {
		t.Fatalf("cycles=%+v err=%v", cycles, err)
	}
}
