package structures

import "time"

type RecentlyAddedMedia struct {
	ID        int       `json:"id"`         // Primary key with auto-increment
	MediaType string    `json:"media_type"` // Type of media (MOVIE, SHOW)
	Title     string    `json:"title"`      // Title of the media
	Year      int       `json:"year"`       // Year the media was released
	Summary   string    `json:"summary"`    // Summary of the media
	IMDBID    string    `json:"imdb_id"`    // IMDb ID of the media
	Poster    string    `json:"poster"`     // URL to the poster of the media
	AddedAt   time.Time `json:"added_at"`   // Time the media was added
}
