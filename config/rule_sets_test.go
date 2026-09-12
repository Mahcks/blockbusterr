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

func TestSaveReplacesConfigWithPrivateMode(t *testing.T) {
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
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %o, want 600", info.Mode().Perm())
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("saved config is empty")
	}
}

func TestSavePreservesStricterModeAndRestoreKeepsRecoveryCopy(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	old := []byte("version: 1.5.0\njobs: {}\nfilters: {}\n")
	if err := os.WriteFile(path, old, 0o400); err != nil {
		t.Fatal(err)
	}
	cfg := &Config{Version: "2.0.0", ConfigFilePath: path}
	if err := cfg.Restore(); err != nil {
		t.Fatal(err)
	}
	for _, check := range []struct {
		path string
		mode os.FileMode
	}{{path, 0o400}, {path + ".backup", 0o600}} {
		info, err := os.Stat(check.path)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != check.mode {
			t.Fatalf("%s mode=%v", check.path, info.Mode().Perm())
		}
	}
	backup, err := os.ReadFile(path + ".backup")
	if err != nil || string(backup) != string(old) {
		t.Fatalf("backup=%q err=%v", backup, err)
	}
}

func TestRestoreBackupFailureLeavesActiveConfigUntouched(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	old := []byte("version: 1.5.0\njobs: {}\nfilters: {}\n")
	if err := os.WriteFile(path, old, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(path+".backup", 0o700); err != nil {
		t.Fatal(err)
	}
	if err := (&Config{Version: "2.0.0", ConfigFilePath: path}).Restore(); err == nil {
		t.Fatal("expected backup failure")
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != string(old) {
		t.Fatalf("active config=%q err=%v", got, err)
	}
}

func TestMigratedGlobalBlocksCanBeRemovedPermanently(t *testing.T) {
	for _, previouslyMigrated := range []bool{false, true} {
		cfg := &Config{ConfigFilePath: filepath.Join(t.TempDir(), "config.yaml")}
		cfg.Filters.Movies.BlacklistedTMDBIds = []int{42}
		cfg.Filters.Shows.BlacklistedTVDBIds = []int{84}
		if previouslyMigrated {
			cfg.RuleSets = []RuleSet{
				{ID: DefaultMoviesRuleSetID, Media: "movie", Movies: &MovieFilters{BlacklistedTMDBIds: []int{42, 99}}},
				{ID: DefaultShowsRuleSetID, Media: "show", Shows: &ShowFilters{BlacklistedTVDBIds: []int{84, 199}}},
			}
		}
		cfg.MigrateRuleSets()
		if len(cfg.TitleExceptions.BlockedMovieTMDBIDs) != 1 || len(cfg.TitleExceptions.BlockedShowTVDBIDs) != 1 {
			t.Fatal("legacy blocks were not migrated")
		}
		cfg.TitleExceptions = TitleExceptions{}
		if err := cfg.Save(); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(cfg.ConfigFilePath)
		if err != nil {
			t.Fatal(err)
		}
		reloaded, err := Parse(data, cfg.ConfigFilePath)
		if err != nil {
			t.Fatal(err)
		}
		if len(reloaded.TitleExceptions.BlockedMovieTMDBIDs)+len(reloaded.TitleExceptions.BlockedShowTVDBIDs) != 0 {
			t.Fatal("removed title exceptions returned after reload")
		}
		_, movies, err := reloaded.ResolveRuleSet(DynamicJob{MediaType: "movie"})
		if err != nil {
			t.Fatal(err)
		}
		_, shows, err := reloaded.ResolveRuleSet(DynamicJob{MediaType: "show"})
		if err != nil {
			t.Fatal(err)
		}
		want := 0
		if previouslyMigrated {
			want = 1
		}
		if len(movies.Movies.BlacklistedTMDBIds) != want || len(shows.Shows.BlacklistedTVDBIds) != want {
			t.Fatal("default rules retained migrated blocks or lost their own blocks")
		}
		if previouslyMigrated && (movies.Movies.BlacklistedTMDBIds[0] != 99 || shows.Shows.BlacklistedTVDBIds[0] != 199) {
			t.Fatal("default rules lost their own blocked IDs")
		}
	}
}
