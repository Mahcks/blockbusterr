package integrations

import (
	"strconv"
	"strings"
)

// Movie is the provider-neutral movie shape used by discovery, filtering, and jobs.
type Movie struct {
	Title          string          `json:"title"`
	Year           int             `json:"year"`
	IDs            IDs             `json:"ids"`
	Genres         []string        `json:"genres"`
	Language       string          `json:"language"`
	Country        string          `json:"country"`
	Runtime        int             `json:"runtime"`
	Overview       string          `json:"overview"`
	Rating         float64         `json:"rating"`
	Votes          int             `json:"votes"`
	Certifications []Certification `json:"certifications,omitempty"`
}

// Show is the provider-neutral TV show shape used by discovery, filtering, and jobs.
type Show struct {
	Title          string          `json:"title"`
	Year           int             `json:"year"`
	IDs            IDs             `json:"ids"`
	Genres         []string        `json:"genres"`
	Language       string          `json:"language"`
	Country        string          `json:"country"`
	Runtime        int             `json:"runtime"`
	Network        string          `json:"network"`
	Overview       string          `json:"overview"`
	Rating         float64         `json:"rating"`
	Votes          int             `json:"votes"`
	Certifications []Certification `json:"certifications,omitempty"`
}

type Certification struct {
	Value   string `json:"value"`
	Country string `json:"country"`
	Source  string `json:"source"`
}

// IDs contains identifiers shared by supported discovery providers and downstream services.
type IDs struct {
	Trakt int    `json:"trakt"`
	Slug  string `json:"slug"`
	IMDB  string `json:"imdb"`
	TMDB  int    `json:"tmdb"`
	TVDB  int    `json:"tvdb"`
	Simkl int    `json:"simkl_id"`
}

func MovieKey(ids IDs) string {
	if ids.TMDB > 0 {
		return "movie:tmdb:" + strconv.Itoa(ids.TMDB)
	}
	if id := strings.TrimSpace(ids.IMDB); id != "" {
		return "movie:imdb:" + id
	}
	return ""
}

func ShowKey(ids IDs) string {
	if ids.TMDB > 0 {
		return "show:tmdb:" + strconv.Itoa(ids.TMDB)
	}
	if ids.TVDB > 0 {
		return "show:tvdb:" + strconv.Itoa(ids.TVDB)
	}
	if id := strings.TrimSpace(ids.IMDB); id != "" {
		return "show:imdb:" + id
	}
	return ""
}

func listMediaCount(mediaType string, movies, shows int) int {
	switch mediaType {
	case "movie":
		return movies
	case "show":
		return shows
	default:
		return movies + shows
	}
}
