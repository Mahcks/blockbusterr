package trakt

import (
	"fmt"

	"github.com/dghubble/sling"
	"github.com/mahcks/blockbusterr/internal/global"
)

func Setup(gctx global.Context) (Service, error) {
	svc := &traktService{
		gctx: gctx,
	}

	svc.base = sling.New().Base("https://api.trakt.tv").
		Set("Content-Type", "application/json").
		Set("trakt-api-version", "2")

	if svc.base == nil {
		return nil, fmt.Errorf("failed to initialize base")
	}

	return svc, nil
}
