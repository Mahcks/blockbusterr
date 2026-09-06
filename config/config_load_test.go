package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewLoadsYAMLAndEnvironmentOverrides(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("tmdb:\n  api_key: from-file\nscoring:\n  enabled: true\njobs:\n  global_limit_movies: 2\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CONFIG_PATH", dir)
	t.Setenv("TMDB_API_KEY", "from-environment")
	t.Setenv("SCORING_ENABLED", "false")
	t.Setenv("JOBS_GLOBAL_LIMIT_MOVIES", "7")

	cfg, err := New("prod")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ConfigFilePath != path || cfg.TMDB.APIKey != "from-environment" || cfg.Scoring.Enabled || cfg.Jobs.GlobalLimitMovies != 7 {
		t.Fatalf("unexpected config: path=%q tmdb=%q scoring=%t movie_limit=%d", cfg.ConfigFilePath, cfg.TMDB.APIKey, cfg.Scoring.Enabled, cfg.Jobs.GlobalLimitMovies)
	}
}
