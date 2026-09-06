package database

import (
	"testing"
	"time"
)

func TestQueryActivityByReasonFiltersClassifiesAndPaginates(t *testing.T) {
	db, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	base := time.Now().Add(-time.Hour)
	entries := []ActivityLog{
		// Two "Minimum Year" rejections for the same title/tmdb id (should
		// group into one row with Count=2), plus one for a different title.
		{Timestamp: base, JobID: "job-a", JobType: "trending_movies", MediaType: "movie", Title: "Old Movie", Year: 1999, TMDBID: 1, Status: "rejected", FilterDetails: `[{"name":"Minimum Year","passed":false,"message":"Year 1999 before minimum 2000"}]`},
		{Timestamp: base.Add(time.Minute), JobID: "job-a", JobType: "trending_movies", MediaType: "movie", Title: "Old Movie", Year: 1999, TMDBID: 1, Status: "rejected", FilterDetails: `[{"name":"Minimum Year","passed":false,"message":"Year 1999 before minimum 2000"}]`},
		{Timestamp: base.Add(2 * time.Minute), JobID: "job-a", JobType: "trending_movies", MediaType: "movie", Title: "Ancient Movie", Year: 1950, TMDBID: 2, Status: "rejected", FilterDetails: `[{"name":"Minimum Year","passed":false,"message":"Year 1950 before minimum 2000"}]`},
		// A different reason entirely; must never show up in a "Minimum Year" filter.
		{Timestamp: base.Add(3 * time.Minute), JobID: "job-a", JobType: "trending_movies", MediaType: "movie", Title: "Bad Genre Movie", Year: 2020, TMDBID: 3, Status: "rejected", FilterDetails: `[{"name":"Blocked genre","passed":false,"message":"Blocked genre matched: Horror"}]`},
		// Not rejected at all; must never show up regardless of its message content.
		{Timestamp: base.Add(4 * time.Minute), JobID: "job-a", JobType: "trending_movies", MediaType: "movie", Title: "Added Movie", Year: 2020, TMDBID: 4, Status: "added", Message: "Year 1999 before minimum 2000"},
	}
	for _, e := range entries {
		if err := db.LogActivity(e); err != nil {
			t.Fatalf("LogActivity() error = %v", err)
		}
	}

	page, err := db.QueryActivity(ActivityQuery{Reason: "Minimum Year", Dedupe: true, Limit: 50})
	if err != nil {
		t.Fatalf("QueryActivity() error = %v", err)
	}
	if page.Total != 2 {
		t.Fatalf("Total = %d, want 2 (Old Movie grouped + Ancient Movie)", page.Total)
	}
	if len(page.Groups) != 2 {
		t.Fatalf("len(Groups) = %d, want 2", len(page.Groups))
	}

	var oldMovie *ActivityLogGroup
	for i := range page.Groups {
		if page.Groups[i].Log.Title == "Old Movie" {
			oldMovie = &page.Groups[i]
		}
		if page.Groups[i].Log.Title == "Bad Genre Movie" || page.Groups[i].Log.Title == "Added Movie" {
			t.Fatalf("unexpected title %q leaked into Minimum Year reason filter", page.Groups[i].Log.Title)
		}
	}
	if oldMovie == nil {
		t.Fatal("Old Movie group not found")
	}
	if oldMovie.Count != 2 || len(oldMovie.History) != 2 {
		t.Fatalf("Old Movie Count=%d len(History)=%d, want 2 and 2", oldMovie.Count, len(oldMovie.History))
	}

	// Sorting by title (ascending) should order "Ancient Movie" before "Old Movie".
	sorted, err := db.QueryActivity(ActivityQuery{Reason: "Minimum Year", Dedupe: true, Sort: "title", Order: "asc", Limit: 50})
	if err != nil {
		t.Fatalf("QueryActivity() sorted error = %v", err)
	}
	if len(sorted.Groups) != 2 || sorted.Groups[0].Log.Title != "Ancient Movie" || sorted.Groups[1].Log.Title != "Old Movie" {
		t.Fatalf("title-sorted groups = %+v, want [Ancient Movie, Old Movie]", sorted.Groups)
	}

	// Pagination: pageSize=1 should return exactly one group and preserve Total.
	paged, err := db.QueryActivity(ActivityQuery{Reason: "Minimum Year", Dedupe: true, Sort: "title", Order: "asc", Limit: 1, Offset: 1})
	if err != nil {
		t.Fatalf("QueryActivity() paged error = %v", err)
	}
	if paged.Total != 2 || len(paged.Groups) != 1 || paged.Groups[0].Log.Title != "Old Movie" {
		t.Fatalf("paged result = %+v, want Total=2 with second page = Old Movie", paged)
	}

	// A status filter that contradicts Reason (only rejected items have one)
	// must be overridden rather than silently returning nothing.
	overridden, err := db.QueryActivity(ActivityQuery{Status: "added", Reason: "Minimum Year", Dedupe: true, Limit: 50})
	if err != nil {
		t.Fatalf("QueryActivity() overridden-status error = %v", err)
	}
	if overridden.Total != 2 {
		t.Fatalf("overridden-status Total = %d, want 2 (Reason should force status=rejected)", overridden.Total)
	}
}
