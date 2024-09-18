package omdb

import (
	"github.com/dghubble/sling"
	"github.com/mahcks/blockbusterr/internal/global"
)

func Setup(gctx global.Context) (Service, error) {
	// Initialize the OMDb service with sling base configuration
	svc := &omdbService{
		gctx: gctx,
		base: sling.New().Base("https://www.omdbapi.com").
			Set("Content-Type", "application/json"),
	}

	return svc, nil
}
