package db

import "database/sql"

// Queries struct to hold the database connection
type Queries struct {
	db *sql.DB
}

// NewQueries initializes a new Queries struct
func NewQueries(db *sql.DB) *Queries {
	return &Queries{db: db}
}
