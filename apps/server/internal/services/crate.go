package services

import (
	"github.com/mahcks/blockbusterr/internal/services/sqlite"
	"github.com/mahcks/blockbusterr/internal/services/trakt"
)

type Crate struct {
	Trakt trakt.Service
	SQL   sqlite.Service
}
