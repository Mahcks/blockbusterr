package jobs

import (
	"context"

	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/integrations"
)

func enrichMovieCertifications(ctx context.Context, cfg *config.Config, movies []integrations.Movie) error {
	rules := cfg.Filters.Movies
	if cfg.TMDB.APIKey == "" || len(rules.AllowedCertifications)+len(rules.BlockedCertifications) == 0 {
		return nil
	}
	return integrations.NewTMDB(integrations.TMDBConfig{APIKey: cfg.TMDB.APIKey}).EnrichMovieCertifications(ctx, movies)
}

func enrichShowCertifications(ctx context.Context, cfg *config.Config, shows []integrations.Show) error {
	rules := cfg.Filters.Shows
	if cfg.TMDB.APIKey == "" || len(rules.AllowedCertifications)+len(rules.BlockedCertifications) == 0 {
		return nil
	}
	return integrations.NewTMDB(integrations.TMDBConfig{APIKey: cfg.TMDB.APIKey}).EnrichShowCertifications(ctx, shows)
}
