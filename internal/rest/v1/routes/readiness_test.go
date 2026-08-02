package routes

import (
	"testing"

	"github.com/mahcks/blockbusterr/config"
)

func TestAssessReadiness(t *testing.T) {
	t.Run("empty configuration needs setup", func(t *testing.T) {
		got := assessReadiness(&config.Config{})
		if got.Complete || got.Discovery.State != readinessNotStarted || got.Delivery.State != readinessNotStarted || got.Automation.State != readinessNotStarted {
			t.Fatalf("unexpected readiness: %#v", got)
		}
	})

	t.Run("valid movie automation is ready when its required services are ready", func(t *testing.T) {
		cfg := &config.Config{}
		cfg.TMDB.APIKey = "tmdb"
		cfg.Radarr.URL, cfg.Radarr.APIKey = "http://radarr", "key"
		cfg.Sonarr.URL, cfg.Sonarr.APIKey = "http://sonarr", "key"
		cfg.Jobs.Mode = "direct"
		cfg.MigrateRuleSets()
		cfg.Jobs.List = []config.DynamicJob{{ID: "movie", Name: "Movies", Enabled: true, Source: "tmdb", MediaType: "movie", RuleSetID: config.DefaultMoviesRuleSetID}}
		got := assessReadiness(cfg)
		if !got.Complete || got.Automation.State != readinessReady {
			t.Fatalf("unexpected readiness: %#v", got)
		}
	})

	t.Run("enabled job reports missing source", func(t *testing.T) {
		cfg := &config.Config{}
		cfg.Radarr.URL, cfg.Radarr.APIKey = "http://radarr", "key"
		cfg.Jobs.List = []config.DynamicJob{{ID: "movie", Name: "Movies", Enabled: true, Source: "tmdb", MediaType: "movie", RuleSetID: config.DefaultMoviesRuleSetID}}
		cfg.MigrateRuleSets()
		got := assessReadiness(cfg)
		if got.Automation.State != readinessAttention || got.Complete {
			t.Fatalf("unexpected readiness: %#v", got)
		}
	})
}
