package jobs

import (
	"testing"

	"github.com/mahcks/blockbusterr/internal/integrations"
)

func TestMatchingSonarrSeriesRejectsUnrelatedFirstResult(t *testing.T) {
	show := integrations.Show{Title: "Fallout", Year: 2024, IDs: integrations.IDs{TVDB: 416744, TMDB: 106379}}
	results := []integrations.SonarrSeries{
		{ID: 107, Title: "Trakt: The Series", Year: 2019, TvdbID: 123},
		{Title: "Fallout", Year: 2024, TvdbID: 416744, TmdbID: 106379},
	}

	series, ok := matchingSonarrSeries(show, results)
	if !ok || series.TvdbID != show.IDs.TVDB || series.ID != 0 {
		t.Fatalf("matched %+v, ok=%t", series, ok)
	}
}

func TestMatchingSonarrSeriesRequiresIdentityOrTitleAndYear(t *testing.T) {
	show := integrations.Show{Title: "The Pitt", Year: 2025, IDs: integrations.IDs{TVDB: 449139}}
	results := []integrations.SonarrSeries{{ID: 107, Title: "Trakt: The Series", Year: 2019, TvdbID: 123}}

	if series, ok := matchingSonarrSeries(show, results); ok {
		t.Fatalf("unexpected match: %+v", series)
	}
}
