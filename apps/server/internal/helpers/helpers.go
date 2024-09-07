package helpers

import (
	"github.com/mahcks/blockbusterr/internal/global"
	"github.com/mahcks/blockbusterr/internal/helpers/ombi"
	"github.com/mahcks/blockbusterr/internal/helpers/radarr"
	"github.com/mahcks/blockbusterr/internal/helpers/trakt"
)

type Helpers struct {
	Trakt  trakt.Service
	Ombi   ombi.Service
	Radarr radarr.Service
}

// Initialize the helpers struct, setting up Trakt service
func SetupHelpers(gctx global.Context) (*Helpers, error) {
	traktService, err := trakt.Setup(gctx)
	if err != nil {
		return nil, err
	}

	ombiService, err := ombi.Setup(gctx)
	if err != nil {
		return nil, err
	}

	radarrService, err := radarr.Setup(gctx)
	if err != nil {
		return nil, err
	}

	return &Helpers{
		Trakt:  traktService,
		Ombi:   ombiService,
		Radarr: radarrService,
	}, nil
}
