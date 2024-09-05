package sqlite

import (
	"database/sql"

	"github.com/mahcks/blockbusterr/internal/db"
)

type Service interface {
	DB() *sql.DB
	Queries() *db.Queries
}

type sqliteService struct {
	db      *sql.DB
	queries *db.Queries
}

func (s *sqliteService) DB() *sql.DB {
	return s.db
}

func (s *sqliteService) Queries() *db.Queries {
	return s.queries
}
