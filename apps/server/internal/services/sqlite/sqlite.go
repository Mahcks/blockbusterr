package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/mahcks/blockbusterr/internal/db"
)

func Setup(ctx context.Context, Version string) (Service, error) {
	svc := &sqliteService{}
	var err error

	if Version == "dev" {
		svc.db, err = sql.Open("sqlite3", "blockbusterr.db")
		if err != nil {
			log.Error("Error opening SQLite database", "error", err)
			return nil, err
		}
	} else {
		svc.db, err = sql.Open("sqlite3", "/app/data/settings.db")
		if err != nil {
			log.Error("Error opening SQLite database", "error", err)
			return nil, err
		}
	}

	log.Info("SQLite database opened")

	err = svc.db.Ping()
	if err != nil {
		log.Error("Error pinging SQLite database", "error", err)
		return nil, err
	}

	log.Info("SQLite database pinged")

	// Initialize the queries
	svc.queries = db.NewQueries(svc.db)

	// Check if queries are properly initialized
	if svc.queries == nil {
		log.Error("Failed to initialize queries")
		return nil, fmt.Errorf("queries initialization failed")
	}

	go func() {
		<-ctx.Done()
		svc.db.Close()
		log.Info("SQLite database closed")
	}()

	return svc, nil
}
