package helpers

import (
	"github.com/mahcks/blockbusterr/internal/global"
	"github.com/mahcks/blockbusterr/internal/helpers/trakt"
)

type Helpers struct {
	Trakt trakt.Service
}

// Initialize the helpers struct, setting up Trakt service
func SetupHelpers(gctx global.Context) (*Helpers, error) {
	traktService, err := trakt.Setup(gctx)
	if err != nil {
		return nil, err
	}

	return &Helpers{
		Trakt: traktService,
	}, nil
}
