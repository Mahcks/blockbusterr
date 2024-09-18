package omdb

import (
	"fmt"

	"github.com/dghubble/sling"
	"github.com/mahcks/blockbusterr/internal/global"
)

func Setup(gctx global.Context) (Service, error) {
	svc := &omdbService{
		gctx: gctx,
	}

	svc.base = sling.New().Base("https://www.omdbapi.com").
		Set("Content-Type", "application/json")

	if svc.base == nil {
		return nil, fmt.Errorf("failed to initialize base")
	}

	return svc, nil
}
