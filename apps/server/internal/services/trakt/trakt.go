package trakt

import (
	"context"

	"github.com/dghubble/sling"
)

type SetupOptions struct {
	ClientID     string
	ClientSecret string
}

func Setup(ctx context.Context, opts SetupOptions) (Service, error) {
	svc := &traktService{}

	svc.base = sling.New().Base("https://api.trakt.tv").
		Set("Content-Type", "application/json").
		Set("trakt-api-version", "2").
		Set("trakt-api-key", opts.ClientID)

	return svc, nil
}
