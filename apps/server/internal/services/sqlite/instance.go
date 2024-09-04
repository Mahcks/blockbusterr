package sqlite

import "database/sql"

type Service interface {
	DB() *sql.DB
}

type sqliteService struct {
	db *sql.DB
}

func (s *sqliteService) DB() *sql.DB {
	return s.db
}
