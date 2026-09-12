package config

import (
	"bytes"
	"math"
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

func TestNewRejectsNonFiniteConfiguration(t *testing.T) {
	for _, value := range []string{".nan", ".inf", "-.inf"} {
		dir := t.TempDir()
		data := []byte("scoring:\n  rating_weight: " + value + "\n")
		if err := os.WriteFile(filepath.Join(dir, "config.yaml"), data, 0o600); err != nil {
			t.Fatal(err)
		}
		t.Setenv("CONFIG_PATH", dir)
		if _, err := New("2.0.0"); err == nil {
			t.Fatalf("startup accepted %s", value)
		}
		if _, err := Parse(data, ""); err == nil {
			t.Fatalf("parser accepted %s", value)
		}
	}
}

func TestNonFiniteValuesCannotBeSavedOrLoadedFromEnvironment(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	original := []byte("version: 2.0.0\njobs: {}\n")
	if err := os.WriteFile(path, original, 0o600); err != nil {
		t.Fatal(err)
	}
	cfg := &Config{ConfigFilePath: path}
	cfg.Jobs.SmartPopularMovies.AdjustmentFactor = math.NaN()
	if err := cfg.Save(); err == nil {
		t.Fatal("save accepted a nonfinite legacy smart-job value")
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(after, original) {
		t.Fatal("failed save changed original configuration")
	}
	t.Setenv("CONFIG_PATH", dir)
	t.Setenv("SCORING_RATING_SCALE", "NaN")
	if _, err := New("2.0.0"); err == nil {
		t.Fatal("startup accepted a nonfinite environment override")
	}
}
