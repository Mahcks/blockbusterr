package database

import (
	"encoding/json"
	"time"

	"github.com/mahcks/blockbusterr/pkg/enums"
)

type SelectionCycleItem struct {
	MediaKey string
	JobID    string
	JobIDs   []string
	Sources  []string
	Score    float64
	Rank     int
	Reason   enums.SelectionReason
	Snapshot any
}

type SelectionCycle struct {
	ID                 int64                      `json:"id"`
	StartedAt          time.Time                  `json:"started_at"`
	FinishedAt         *time.Time                 `json:"finished_at,omitempty"`
	Status             enums.SelectionCycleStatus `json:"status"`
	MovieWinners       int                        `json:"movie_winners"`
	ShowWinners        int                        `json:"show_winners"`
	MovieDelivered     int                        `json:"movie_delivered"`
	ShowDelivered      int                        `json:"show_delivered"`
	FailedItems        int                        `json:"failed_items"`
	AccountingComplete bool                       `json:"accounting_complete"`
	ErrorMessage       string                     `json:"error_message,omitempty"`
}

func (d *Database) StartSelectionCycle(startedAt time.Time) (int64, error) {
	result, err := d.db.Exec("INSERT INTO selection_cycles (started_at, status) VALUES (?, ?)", startedAt, enums.SelectionCycleRunning)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (d *Database) SaveSelectionCycleItems(cycleID int64, items []SelectionCycleItem) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	for _, item := range items {
		jobIDs, _ := json.Marshal(item.JobIDs)
		sources, _ := json.Marshal(item.Sources)
		snapshot, err := json.Marshal(item.Snapshot)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(`INSERT INTO selection_cycle_items (cycle_id, media_key, job_id, job_ids, sources, score, rank, reason, snapshot) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, cycleID, item.MediaKey, item.JobID, string(jobIDs), string(sources), item.Score, item.Rank, item.Reason, string(snapshot)); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (d *Database) CompleteSelectionCycle(cycleID int64, status enums.SelectionCycleStatus, movieWinners, showWinners, movieDelivered, showDelivered, failedItems int, message string) error {
	_, err := d.db.Exec(`UPDATE selection_cycles SET finished_at = ?, status = ?, movie_winners = ?, show_winners = ?, movie_delivered = ?, show_delivered = ?, failed_items = ?, accounting_complete = 1, error_message = ? WHERE id = ?`, time.Now(), status, movieWinners, showWinners, movieDelivered, showDelivered, failedItems, message, cycleID)
	return err
}

func (d *Database) GetRecentSelectionCycles(limit int) ([]SelectionCycle, error) {
	rows, err := d.db.Query(`SELECT id, started_at, finished_at, status, movie_winners, show_winners, movie_delivered, show_delivered, failed_items, accounting_complete, COALESCE(error_message, '') FROM selection_cycles ORDER BY started_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	cycles := []SelectionCycle{}
	for rows.Next() {
		var cycle SelectionCycle
		if err := rows.Scan(&cycle.ID, &cycle.StartedAt, &cycle.FinishedAt, &cycle.Status, &cycle.MovieWinners, &cycle.ShowWinners, &cycle.MovieDelivered, &cycle.ShowDelivered, &cycle.FailedItems, &cycle.AccountingComplete, &cycle.ErrorMessage); err != nil {
			return nil, err
		}
		cycles = append(cycles, cycle)
	}
	return cycles, rows.Err()
}
