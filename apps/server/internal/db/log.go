package db

import (
	"context"
	"fmt"

	"github.com/mahcks/blockbusterr/pkg/structures"
)

func (q *Queries) InsertLog(ctx context.Context, level structures.LogLevel, label, message string) error {
	query := `INSERT INTO logs (level, label, message) VALUES ($1, $2, $3)`

	_, err := q.db.ExecContext(ctx, query, level.String(), label, message)
	if err != nil {
		return fmt.Errorf("error inserting log: %v", err)
	}

	return nil
}

func (q *Queries) GetLogs(ctx context.Context, take, skip int, filter, search string) ([]structures.Log, error) {
	// Base query
	query := `SELECT id, level, label, message, timestamp FROM logs WHERE 1=1`
	args := []interface{}{}
	argIndex := 1

	// Add filter by level if provided
	if filter != "" {
		query += fmt.Sprintf(" AND level = $%d", argIndex)
		args = append(args, filter)
		argIndex++
	}

	// Add search condition if provided
	if search != "" {
		query += fmt.Sprintf(" AND (label LIKE $%d OR message LIKE $%d)", argIndex, argIndex+1)
		args = append(args, "%"+search+"%", "%"+search+"%")
		argIndex += 2
	}

	// Add order, limit, and offset for pagination
	query += fmt.Sprintf(" ORDER BY timestamp DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, take, skip)

	// Prepare the result slice
	var logs []structures.Log

	// Execute the query
	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying logs: %v", err)
	}
	defer rows.Close()

	// Loop over the rows and scan the result into the logs slice
	for rows.Next() {
		var log structures.Log
		if err := rows.Scan(&log.ID, &log.Level, &log.Label, &log.Message, &log.Timestamp); err != nil {
			return nil, fmt.Errorf("error scanning log row: %v", err)
		}
		logs = append(logs, log)
	}

	// Return the result set
	return logs, nil
}
