package config

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestOAuthSecretsAreExcludedFromJSON(t *testing.T) {
	cfg := Config{}
	cfg.Trakt.AccessToken = "trakt-access"
	cfg.Trakt.RefreshToken = "trakt-refresh"
	cfg.TMDB.SessionID = "tmdb-session"

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{"trakt-access", "trakt-refresh", "tmdb-session"} {
		if strings.Contains(string(data), secret) {
			t.Fatalf("JSON exposed OAuth secret %q", secret)
		}
	}
}
