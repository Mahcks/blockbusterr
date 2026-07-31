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
