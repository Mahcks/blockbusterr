package global

import (
	"context"
	"path/filepath"
	"sync"
	"testing"

	"github.com/mahcks/blockbusterr/config"
)

func TestConfigUpdatesAreCopyOnWrite(t *testing.T) {
	cfg := &config.Config{ConfigFilePath: filepath.Join(t.TempDir(), "config.yaml")}
	cfg.Jobs.GlobalLimitMovies = 1
	gctx := New(context.Background(), cfg, nil, "test", "test", nil)

	snapshot := gctx.Config()
	snapshot.Jobs.GlobalLimitMovies = 99
	if got := gctx.Config().Jobs.GlobalLimitMovies; got != 1 {
		t.Fatalf("snapshot mutation changed live config: %d", got)
	}

	var wg sync.WaitGroup
	for range 10 {
		wg.Go(func() {
			if err := UpdateConfig(gctx, func(candidate *config.Config) error {
				candidate.Jobs.GlobalLimitMovies++
				return nil
			}); err != nil {
				t.Errorf("UpdateConfig() error = %v", err)
			}
		})
	}
	wg.Wait()
	if got := gctx.Config().Jobs.GlobalLimitMovies; got != 11 {
		t.Fatalf("serialized updates = %d, want 11", got)
	}
}

func TestConfigUpdateFailureDoesNotPublish(t *testing.T) {
	cfg := &config.Config{ConfigFilePath: filepath.Join(t.TempDir(), "config.yaml")}
	cfg.Jobs.GlobalLimitMovies = 1
	gctx := New(context.Background(), cfg, nil, "test", "test", nil)
	err := UpdateConfig(gctx, func(candidate *config.Config) error {
		candidate.Jobs.GlobalLimitMovies = 99
		candidate.ConfigFilePath = filepath.Join(t.TempDir(), "missing", "config.yaml")
		return nil
	})
	if err == nil {
		t.Fatal("expected save failure")
	}
	if got := gctx.Config().Jobs.GlobalLimitMovies; got != 1 {
		t.Fatalf("failed update published value %d", got)
	}
}
