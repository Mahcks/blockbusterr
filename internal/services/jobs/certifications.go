package jobs

import (
	"context"

	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/integrations"
)

func enrichMovieCertifications(ctx context.Context, cfg *config.Config, movies []integrations.Movie) {
	rules := cfg.Filters.Movies
	if cfg.TMDB.APIKey == "" || len(rules.AllowedCertifications)+len(rules.BlockedCertifications) == 0 {
		return
	}
	integrations.NewTMDB(integrations.TMDBConfig{APIKey: cfg.TMDB.APIKey}).EnrichMovieCertifications(ctx, movies)
}

func enrichShowCertifications(ctx context.Context, cfg *config.Config, shows []integrations.Show) {
	rules := cfg.Filters.Shows
	if cfg.TMDB.APIKey == "" || len(rules.AllowedCertifications)+len(rules.BlockedCertifications) == 0 {
		return
	}
	integrations.NewTMDB(integrations.TMDBConfig{APIKey: cfg.TMDB.APIKey}).EnrichShowCertifications(ctx, shows)
}
