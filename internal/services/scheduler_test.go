package services

import (
	"testing"

	"github.com/mahcks/blockbusterr/config"
)

func TestRunnableJobsSkipsUnconfiguredProviders(t *testing.T) {
	cfg := &config.Config{}
	cfg.Simkl.ClientID = "configured"
	cfg.Jobs.List = []config.DynamicJob{
		{ID: "simkl", Enabled: true, Source: "simkl"},
		{ID: "trakt", Enabled: true, Source: "trakt"},
		{ID: "disabled", Enabled: false, Source: "simkl"},
	}

	jobs := runnableJobs(cfg)
	if len(jobs) != 1 || jobs[0].ID != "simkl" {
		t.Fatalf("runnableJobs() = %#v, want only Simkl job", jobs)
	}
}

func TestSplitSelectionJobs(t *testing.T) {
	cfg := &config.Config{}
	cfg.Jobs.Selection.Enabled = true
	cycle, standalone := splitSelectionJobs(cfg, []config.DynamicJob{{ID: "ranked", SelectionCycle: true}, {ID: "normal"}})
	if len(cycle) != 1 || cycle[0].ID != "ranked" || len(standalone) != 1 || standalone[0].ID != "normal" {
		t.Fatalf("cycle=%+v standalone=%+v", cycle, standalone)
	}
}

func TestRunsAtStartupOnlyForDurationSchedules(t *testing.T) {
	if !runsAtStartup("24h") {
		t.Fatal("duration schedule should run at startup")
	}
	if runsAtStartup("0 8 * * *") {
		t.Fatal("cron schedule should wait for its next occurrence")
	}
}

func TestJobSignatureIncludesExecutionSettings(t *testing.T) {
	base := config.DynamicJob{ID: "job", Name: "Job", Source: "tmdb", MediaType: "movie"}
	changed := base
	changed.DeliveryLimit = 5
	if jobSignature(base) == jobSignature(changed) {
		t.Fatal("delivery limit must change the scheduler signature")
	}
	changed = base
	changed.RepeatPolicy = "never"
	if jobSignature(base) == jobSignature(changed) {
		t.Fatal("repeat policy must change the scheduler signature")
	}
}
