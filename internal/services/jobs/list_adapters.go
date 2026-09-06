package jobs

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/integrations"
	"github.com/mahcks/blockbusterr/pkg/enums"
)

type traktListSource struct{ client *integrations.Trakt }

func (source traktListSource) FetchList(ctx context.Context, locator config.ListLocator, mediaType string, limit int) (ListResult, error) {
	listID := locator.ListID
	if listID == "" {
		listID = locator.Slug
	}
	items, err := source.client.GetListItems(ctx, locator.Owner, listID, enums.ListKind(locator.Kind) == enums.ListKindWatchlist, mediaType, limit)
	return ListResult{Source: "trakt", Name: listID, Movies: items.Movies, Shows: items.Shows}, err
}

type tmdbListSource struct{ client *integrations.TMDB }

func (source tmdbListSource) FetchList(ctx context.Context, locator config.ListLocator, mediaType string, limit int) (ListResult, error) {
	listID := locator.ListID
	if listID == "" {
		listID = locator.Slug
	}
	items, err := source.client.GetListItems(ctx, listID, enums.ListKind(locator.Kind) == enums.ListKindWatchlist, mediaType, limit)
	return ListResult{Source: "tmdb", Name: items.Name, Movies: items.Movies, Shows: items.Shows}, err
}

type mdbListSource struct{ client *integrations.MDBList }

func (source mdbListSource) FetchList(ctx context.Context, locator config.ListLocator, mediaType string, limit int) (ListResult, error) {
	listID := locator.ListID
	if listID == "" {
		listID = locator.Slug
	}
	items, err := source.client.GetListItems(ctx, locator.Owner, listID, enums.ListKind(locator.Kind) == enums.ListKindWatchlist, mediaType, limit)
	return ListResult{Source: "mdblist", Name: items.Name, Movies: items.Movies, Shows: items.Shows}, err
}

type letterboxdListSource struct{ client *integrations.Letterboxd }

func (source letterboxdListSource) FetchList(ctx context.Context, locator config.ListLocator, mediaType string, limit int) (ListResult, error) {
	listID := locator.ListID
	if listID == "" {
		listID = locator.Slug
	}
	items, err := source.client.GetListItems(ctx, locator.Owner, listID, enums.ListKind(locator.Kind) == enums.ListKindWatchlist, mediaType, limit)
	for _, warning := range items.Warnings {
		slog.Warn("Letterboxd item skipped", "warning", warning)
	}
	return ListResult{Source: "letterboxd", Name: items.Name, Movies: items.Movies, Shows: items.Shows, Warnings: items.Warnings}, err
}

func init() {
	RegisterListSource("trakt", func(cfg *config.Config) (ListSource, error) {
		if cfg.Trakt.ClientID == "" {
			return nil, fmt.Errorf("trakt client ID is not configured")
		}
		client := integrations.NewTrakt(integrations.TraktConfig{ClientID: cfg.Trakt.ClientID, ClientSecret: cfg.Trakt.ClientSecret, AccessToken: cfg.Trakt.AccessToken, RefreshToken: cfg.Trakt.RefreshToken, TokenExpires: cfg.Trakt.TokenExpires, LoadToken: func() integrations.TraktToken {
			access, refresh, expires := cfg.Trakt.AccessToken, cfg.Trakt.RefreshToken, cfg.Trakt.TokenExpires
			if cfg.LoadTraktToken != nil {
				access, refresh, expires = cfg.LoadTraktToken()
			}
			return integrations.TraktToken{AccessToken: access, RefreshToken: refresh, CreatedAt: expires, ExpiresIn: 0}
		}, OnToken: func(token integrations.TraktToken) error {
			if cfg.UpdateTraktToken != nil {
				return cfg.UpdateTraktToken(token.AccessToken, token.RefreshToken, token.ExpiresAt())
			}
			cfg.Trakt.AccessToken, cfg.Trakt.RefreshToken, cfg.Trakt.TokenExpires = token.AccessToken, token.RefreshToken, token.ExpiresAt()
			return cfg.Save()
		}})
		return traktListSource{client: client}, nil
	})
	RegisterListSource("tmdb", func(cfg *config.Config) (ListSource, error) {
		if cfg.TMDB.APIKey == "" {
			return nil, fmt.Errorf("TMDB API key is not configured")
		}
		return tmdbListSource{client: integrations.NewTMDB(integrations.TMDBConfig{APIKey: cfg.TMDB.APIKey, SessionID: cfg.TMDB.SessionID, AccountID: cfg.TMDB.AccountID})}, nil
	})
	RegisterListSource("mdblist", func(cfg *config.Config) (ListSource, error) {
		if cfg.MDBList.APIKey == "" {
			return nil, fmt.Errorf("MDBList API key is not configured")
		}
		return mdbListSource{client: integrations.NewMDBList(integrations.MDBListConfig{APIKey: cfg.MDBList.APIKey})}, nil
	})
	RegisterListSource("letterboxd", func(cfg *config.Config) (ListSource, error) {
		if !cfg.Letterboxd.ExperimentalScraping {
			return nil, fmt.Errorf("letterboxd experimental scraping is disabled")
		}
		return letterboxdListSource{client: integrations.NewLetterboxd(integrations.LetterboxdConfig{TMDBAPIKey: cfg.TMDB.APIKey})}, nil
	})
}
