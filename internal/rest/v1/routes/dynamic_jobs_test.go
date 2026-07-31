package routes

import (
	"testing"

	"github.com/mahcks/blockbusterr/config"
)

func TestValidateDynamicJobSources(t *testing.T) {
	cfg := &config.Config{}
	cfg.Trakt.ClientID = "configured"
	cfg.TMDB.APIKey = "configured"
	cfg.Simkl.ClientID = "configured"
	tests := []struct {
		name    string
		job     config.DynamicJob
		wantErr bool
	}{
		{name: "TMDB trending", job: config.DynamicJob{Name: "Trending", Type: "trending", Source: "tmdb", MediaType: "movie", Limit: 50}},
		{name: "Simkl watched", job: config.DynamicJob{Name: "Watched", Type: "watched", Source: "simkl", MediaType: "show", Limit: 50, Period: "weekly"}},
		{name: "TMDB collected unsupported", job: config.DynamicJob{Name: "Collected", Type: "collected", Source: "tmdb", MediaType: "movie", Limit: 50, Period: "weekly"}, wantErr: true},
		{name: "Simkl yearly unsupported", job: config.DynamicJob{Name: "Watched", Type: "watched", Source: "simkl", MediaType: "movie", Limit: 50, Period: "yearly"}, wantErr: true},
		{name: "Simkl limit", job: config.DynamicJob{Name: "Trending", Type: "trending", Source: "simkl", MediaType: "movie", Limit: 501}, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateDynamicJob(cfg, test.job)
			if (err != nil) != test.wantErr {
				t.Fatalf("validateDynamicJob() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

func TestValidateDynamicJobRejectsUnconfiguredProvider(t *testing.T) {
	job := config.DynamicJob{Name: "Trending", Type: "trending", Source: "simkl", MediaType: "movie", Limit: 50}
	if err := validateDynamicJob(&config.Config{}, job); err == nil {
		t.Fatal("expected unconfigured provider error")
	}
}
