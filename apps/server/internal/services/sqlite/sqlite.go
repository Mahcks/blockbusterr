package sqlite

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/mahcks/blockbusterr/internal/db"
)

func Setup(ctx context.Context) (Service, error) {
	svc := &sqliteService{}
	var err error

	svc.db, err = sql.Open("sqlite3", "blockbusterr.db")
	if err != nil {
		slog.Error("Error opening SQLite database", "error", err)
		return nil, err
	}

	slog.Info("SQLite database opened")

	err = svc.db.Ping()
	if err != nil {
		slog.Error("Error pinging SQLite database", "error", err)
		return nil, err
	}

	slog.Info("SQLite database pinged")

	// Initialize the queries
	svc.queries = db.NewQueries(svc.db)

	go func() {
		<-ctx.Done()
		svc.db.Close()
		slog.Info("SQLite database closed")
	}()

	return svc, nil
}
