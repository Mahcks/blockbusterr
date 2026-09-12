package integrations

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const LetterboxdBaseURL = "https://letterboxd.com"

var (
	letterboxdItemPattern  = regexp.MustCompile(`(?s)<li[^>]*(?:posteritem|griditem)[^>]*>(.*?)</li>`)
	letterboxdLinkPattern  = regexp.MustCompile(`data-target-link="([^"]+)"`)
	letterboxdTitlePattern = regexp.MustCompile(`data-item-name="([^"]+)"`)
	letterboxdYearPattern  = regexp.MustCompile(`data-item-full-display-name="[^"]*\((\d{4})\)"`)
	letterboxdTMDBPattern  = regexp.MustCompile(`themoviedb\.org/(movie|tv)/(\d+)`)
	letterboxdNextPattern  = regexp.MustCompile(`(?i)(?:rel="next"[^>]*href="([^"]+)"|href="([^"]+)"[^>]*rel="next")`)
)

type Letterboxd struct {
	httpClient *http.Client
	tmdb       *TMDB
}

type LetterboxdConfig struct{ TMDBAPIKey string }

type LetterboxdListItems struct {
	Name     string
	Movies   []Movie
	Shows    []Show
	Warnings []string
}

type letterboxdItem struct {
	Title string
	Year  int
	Path  string
}

func NewLetterboxd(config LetterboxdConfig) *Letterboxd {
	return &Letterboxd{httpClient: &http.Client{Timeout: 20 * time.Second}, tmdb: NewTMDB(TMDBConfig{APIKey: config.TMDBAPIKey})}
}

func (client *Letterboxd) GetListItems(ctx context.Context, owner, identifier string, watchlist bool, mediaType string, limit int) (LetterboxdListItems, error) {
	if owner == "" {
		return LetterboxdListItems{}, fmt.Errorf("Letterboxd requires a public member name")
	}
	path := "/" + url.PathEscape(owner) + "/watchlist/"
	name := owner + "'s watchlist"
	if !watchlist {
		if identifier == "" {
			return LetterboxdListItems{}, fmt.Errorf("Letterboxd requires a public list slug")
		}
		path = "/" + url.PathEscape(owner) + "/list/" + url.PathEscape(identifier) + "/"
		name = identifier
	}

	result := LetterboxdListItems{Name: name}
	for page := 1; listMediaCount(mediaType, len(result.Movies), len(result.Shows)) < limit; page++ {
		pagePath := path
		if page > 1 {
			pagePath += "page/" + strconv.Itoa(page) + "/"
		}
		body, err := client.getHTML(ctx, pagePath, 5<<20)
		if err != nil {
			return LetterboxdListItems{}, err
		}
		items := parseLetterboxdItems(body, -1)
		if len(items) == 0 {
			if page == 1 && !strings.Contains(strings.ToLower(body), "empty") {
				return LetterboxdListItems{}, fmt.Errorf("Letterboxd page structure changed or access was blocked; use MDBList or disable experimental scraping")
			}
			break
		}
		for _, item := range items {
			if listMediaCount(mediaType, len(result.Movies), len(result.Shows)) >= limit {
				break
			}
			filmPage, err := client.getHTML(ctx, item.Path, 2<<20)
			if err != nil {
				if ctx.Err() != nil {
					return LetterboxdListItems{}, ctx.Err()
				}
				result.Warnings = append(result.Warnings, fmt.Sprintf("%s: %v", item.Title, err))
				continue
			}
			match := letterboxdTMDBPattern.FindStringSubmatch(filmPage)
			if len(match) != 3 {
				result.Warnings = append(result.Warnings, fmt.Sprintf("%s: no TMDB ID; skipped unsafe title-only match", item.Title))
				continue
			}
			id, _ := strconv.Atoi(match[2])
			if match[1] == "movie" && mediaType != "show" {
				result.Movies = append(result.Movies, Movie{Title: item.Title, Year: item.Year, IDs: IDs{TMDB: id}})
			} else if match[1] == "tv" && mediaType != "movie" {
				result.Shows = append(result.Shows, Show{Title: item.Title, Year: item.Year, IDs: IDs{TMDB: id}})
			}
		}
		if len(letterboxdNextPattern.FindStringSubmatch(body)) == 0 {
			break
		}
	}
	if client.tmdb.apiKey != "" {
		if err := client.tmdb.enrichShows(ctx, result.Shows); err != nil {
			return LetterboxdListItems{}, err
		}
	}
	return result, nil
}

func parseLetterboxdItems(document string, limit int) []letterboxdItem {
	matches := letterboxdItemPattern.FindAllStringSubmatch(document, limit)
	items := make([]letterboxdItem, 0, len(matches))
	for _, match := range matches {
		link, title := letterboxdLinkPattern.FindStringSubmatch(match[1]), letterboxdTitlePattern.FindStringSubmatch(match[1])
		if len(link) != 2 || len(title) != 2 || !strings.HasPrefix(link[1], "/film/") {
			continue
		}
		year := 0
		if value := letterboxdYearPattern.FindStringSubmatch(match[1]); len(value) == 2 {
			year, _ = strconv.Atoi(value[1])
		}
		items = append(items, letterboxdItem{Title: html.UnescapeString(strings.TrimSpace(strings.TrimSuffix(title[1], fmt.Sprintf(" (%d)", year)))), Year: year, Path: link[1]})
	}
	return items
}

func (client *Letterboxd) getHTML(ctx context.Context, path string, maxBytes int64) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, LetterboxdBaseURL+path, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "text/html")
	req.Header.Set("User-Agent", "Blockbusterr/2 (+https://github.com/Mahcks/blockbusterr)")
	response, err := doRequest(client.httpClient, req)
	if err != nil {
		return "", fmt.Errorf("Letterboxd request failed: %w", err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("Letterboxd list or watchlist was not found or is private")
	}
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Letterboxd returned status %d; experimental scraping may be blocked", response.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, maxBytes+1))
	if err != nil {
		return "", fmt.Errorf("failed to read Letterboxd response: %w", err)
	}
	if int64(len(body)) > maxBytes {
		return "", fmt.Errorf("Letterboxd response exceeded the safety limit")
	}
	return string(body), nil
}
