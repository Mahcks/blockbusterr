package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMigrateRuleSetsIsDeterministic(t *testing.T) {
	cfg := &Config{}
	cfg.Filters.Movies.MinRating = 7
	cfg.Jobs.List = []DynamicJob{{ID: "job-1", Name: "Recent", MediaType: "movie", UseCustomFilters: true, Filters: FilterConfig{Movies: MovieFilters{MinRating: 8, BlacklistedTMDBIds: []int{42}}}}}
	cfg.MigrateRuleSets()
	cfg.MigrateRuleSets()
	if len(cfg.RuleSets) != 3 {
		t.Fatalf("got %d rule sets, want 3", len(cfg.RuleSets))
	}
	if cfg.Jobs.List[0].RuleSetID != "migrated-job-1" {
		t.Fatalf("assignment = %q", cfg.Jobs.List[0].RuleSetID)
	}
	_, effective, err := cfg.ResolveRuleSet(cfg.Jobs.List[0])
	if err != nil {
		t.Fatal(err)
	}
	if effective.Movies.MinRating != 8 {
		t.Fatalf("rating = %v, want 8", effective.Movies.MinRating)
	}
	if len(cfg.TitleExceptions.BlockedMovieTMDBIDs) != 0 {
		t.Fatalf("custom block leaked globally: %v", cfg.TitleExceptions.BlockedMovieTMDBIDs)
	}
	_, defaults, err := cfg.ResolveRuleSet(DynamicJob{MediaType: "movie", RuleSetID: DefaultMoviesRuleSetID})
	if err != nil {
		t.Fatal(err)
	}
	if len(defaults.Movies.BlacklistedTMDBIds) != 0 {
		t.Fatalf("custom block affected defaults: %v", defaults.Movies.BlacklistedTMDBIds)
	}
	if len(effective.Movies.BlacklistedTMDBIds) != 1 || effective.Movies.BlacklistedTMDBIds[0] != 42 {
		t.Fatalf("custom block not preserved: %v", effective.Movies.BlacklistedTMDBIds)
	}
	if effective.Movies.UnknownCertification != "allow" || defaults.Movies.UnknownCertification != "allow" {
		t.Fatalf("unknown certification migration = %q, %q", effective.Movies.UnknownCertification, defaults.Movies.UnknownCertification)
	}
}

func TestValidateCertificationRules(t *testing.T) {
	cfg := &Config{}
	base := RuleSet{ID: "family", Name: "Family", Media: "movie", Movies: &MovieFilters{CertificationCountry: "US", AllowedCertifications: []string{"G", "PG"}, UnknownCertification: "reject"}}
	if err := cfg.ValidateRuleSet(base, ""); err != nil {
		t.Fatal(err)
	}
	base.Movies.CertificationCountry = "USA"
	if err := cfg.ValidateRuleSet(base, ""); err == nil {
		t.Fatal("expected invalid certification country")
	}
	base.Movies.CertificationCountry = "US"
	base.Movies.UnknownCertification = "maybe"
	if err := cfg.ValidateRuleSet(base, ""); err == nil {
		t.Fatal("expected invalid unknown certification policy")
	}
}

func TestResolveRuleSetFailsClosedOnMediaMismatch(t *testing.T) {
	cfg := &Config{}
	cfg.MigrateRuleSets()
	_, _, err := cfg.ResolveRuleSet(DynamicJob{MediaType: "show", RuleSetID: DefaultMoviesRuleSetID})
	if err == nil {
		t.Fatal("expected media mismatch error")
	}
}

func TestSaveReplacesConfigAndPreservesMode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("version: old\n"), 0o640); err != nil {
		t.Fatal(err)
	}
	cfg := &Config{Version: "new", ConfigFilePath: path}
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o640 {
		t.Fatalf("mode = %o, want 640", info.Mode().Perm())
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("saved config is empty")
	}
}
