package integrations

import (
	"encoding/json"
	"testing"
)

func TestSimklItemDecoding(t *testing.T) {
	var item simklItem
	data := []byte(`{"title":"Example","ids":{"simkl_id":42,"tmdb":"123","tvdb":"456","imdb":"tt1"},"release_date":"06/24/2026","runtime":"1h 48m","ratings":{"simkl":{"rating":8.1,"votes":900}}}`)
	if err := json.Unmarshal(data, &item); err != nil {
		t.Fatal(err)
	}
	if item.Title != "Example" || int(item.IDs.TMDB) != 123 || int(item.IDs.TVDB) != 456 {
		t.Fatalf("unexpected item: %+v", item)
	}
	if got := parseSimklRuntime(item.Runtime); got != 108 {
		t.Fatalf("runtime = %d, want 108", got)
	}
	if got := simklYear(item.ReleaseDate); got != 2026 {
		t.Fatalf("year = %d, want 2026", got)
	}
}
