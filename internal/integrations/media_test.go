package integrations

import "testing"

func TestCanonicalKeysUseStableFallbacks(t *testing.T) {
	first := ShowKey(IDs{TMDB: 10})
	second := ShowKey(IDs{TMDB: 20})
	if first == "" || second == "" || first == second {
		t.Fatalf("TMDB-only shows collided: %q %q", first, second)
	}
	if ShowKey(IDs{IMDB: "tt1"}) == "" || MovieKey(IDs{IMDB: "tt2"}) == "" || ShowKey(IDs{}) != "" {
		t.Fatal("stable ID fallback is incorrect")
	}
}
