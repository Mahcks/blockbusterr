package sonarr

import (
	"context"
	"fmt"
	"time"

	"github.com/dghubble/sling"
	"github.com/mahcks/blockbusterr/internal/global"
	"github.com/mahcks/blockbusterr/pkg/errors"
)

type Service interface {
	GetRootFolders() (GetRootFoldersResponse, error)
	GetQualityProfiles() (GetQualityProfilesResponse, error)
	RequestSeries(ctx context.Context, body RequestSeriesBody) (RequestSeriesResponse, error)
}

type sonarrService struct {
	gctx global.Context
}

func (r *sonarrService) FetchSonarrURLFromDB() (*sling.Sling, error) {
	sonarrSettings, err := r.gctx.Crate().SQL.Queries().GetSonarrSettings(r.gctx)
	if err != nil {
		return nil, err
	}

	// Check if Radarr URL is valid and not empty
	if !sonarrSettings.URL.Valid || sonarrSettings.URL.String == "" {
		return nil, errors.ErrInternalServerError().SetDetail("Sonarr URL is set but empty")
	}

	// Check if Radarr API Key is valid and not empty
	if !sonarrSettings.APIKey.Valid || sonarrSettings.APIKey.String == "" {
		return nil, errors.ErrInternalServerError().SetDetail("Sonarr API Key is set but empty")
	}

	base := sling.New().Base(sonarrSettings.URL.String).
		Set("Content-Type", "application/json").
		Set("X-Api-Key", sonarrSettings.APIKey.String)

	// Return the configured Radarr URL
	return base, nil
}

type GetRootFoldersResponse []RootFolder

type RootFolder struct {
	ID              int                        `json:"id"`
	Path            string                     `json:"path"`
	Accessible      bool                       `json:"accessible"`
	FreeSpace       int64                      `json:"freeSpace"`
	UnmappedFolders []RootFolderUnmappedFolder `json:"unmappedFolders"`
}

type RootFolderUnmappedFolder struct {
	Name         string `json:"name"`
	Path         string `json:"path"`
	RelativePath string `json:"relativePath"`
}

func (r *sonarrService) GetRootFolders() (GetRootFoldersResponse, error) {
	url, err := r.FetchSonarrURLFromDB()
	if err != nil {
		return nil, err
	}

	var response GetRootFoldersResponse
	_, err = url.New().Get("/api/v3/rootfolder").Receive(&response, nil)
	if err != nil {
		return nil, errors.ErrInternalServerError().SetDetail("Failed to get Radarr root folders")
	}

	return response, nil
}

type GetQualityProfilesResponse []QualityProfile

type Quality struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Source     string `json:"source"`
	Resolution int    `json:"resolution"`
}

type Item struct {
	Quality *Quality `json:"quality,omitempty"` // Optional field for nested items
	Items   []Item   `json:"items"`
	Allowed bool     `json:"allowed"`
	Name    string   `json:"name,omitempty"` // Optional field for named items
	ID      int      `json:"id,omitempty"`   // Optional field for named items
}

type QualityProfile struct {
	Name              string `json:"name"`
	UpgradeAllowed    bool   `json:"upgradeAllowed"`
	Cutoff            int    `json:"cutoff"`
	Items             []Item `json:"items"`
	MinFormatScore    int    `json:"minFormatScore"`
	CutoffFormatScore int    `json:"cutoffFormatScore"`
	FormatItems       []Item `json:"formatItems"`
	ID                int    `json:"id"`
}

func (r *sonarrService) GetQualityProfiles() (GetQualityProfilesResponse, error) {
	url, err := r.FetchSonarrURLFromDB()
	if err != nil {
		return nil, err
	}

	var response GetQualityProfilesResponse
	_, err = url.New().Get("/api/v3/qualityprofile").Receive(&response, nil)
	if err != nil {
		return nil, errors.ErrInternalServerError().SetDetail("Failed to get Radarr quality profiles")
	}

	return response, nil
}

// Series represents a TV series with various details
type Series struct {
	Title             string           `json:"title"`
	AlternateTitles   []AlternateTitle `json:"alternateTitles"`
	SortTitle         string           `json:"sortTitle"`
	Status            string           `json:"status"`
	Ended             bool             `json:"ended"`
	Overview          string           `json:"overview"`
	Network           string           `json:"network"`
	AirTime           string           `json:"airTime"`
	Images            []Image          `json:"images"`
	OriginalLanguage  Language         `json:"originalLanguage"`
	Seasons           []Season         `json:"seasons"`
	Year              int              `json:"year"`
	Path              string           `json:"path"`
	QualityProfileId  int              `json:"qualityProfileId"`
	SeasonFolder      bool             `json:"seasonFolder"`
	Monitored         bool             `json:"monitored"`
	MonitorNewItems   string           `json:"monitorNewItems"`
	UseSceneNumbering bool             `json:"useSceneNumbering"`
	Runtime           int              `json:"runtime"`
	TvdbId            int              `json:"tvdbId"`
	TvRageId          int              `json:"tvRageId"`
	TvMazeId          int              `json:"tvMazeId"`
	TmdbId            int              `json:"tmdbId"`
	FirstAired        time.Time        `json:"firstAired"`
	LastAired         time.Time        `json:"lastAired"`
	SeriesType        string           `json:"seriesType"`
	CleanTitle        string           `json:"cleanTitle"`
	ImdbId            string           `json:"imdbId"`
	TitleSlug         string           `json:"titleSlug"`
	RootFolderPath    string           `json:"rootFolderPath"`
	Certification     string           `json:"certification"`
	Genres            []string         `json:"genres"`
	Tags              []string         `json:"tags"`
	Added             time.Time        `json:"added"`
	AddOptions        AddOptions       `json:"addOptions"`
	Ratings           Ratings          `json:"ratings"`
	Statistics        Statistics       `json:"statistics"`
	LanguageProfileId int              `json:"languageProfileId"`
	Id                int              `json:"id"`
}

// AlternateTitle represents an alternate title for the series
type AlternateTitle struct {
	Title        string `json:"title"`
	SeasonNumber int    `json:"seasonNumber"`
	Comment      string `json:"comment,omitempty"` // Optional field
}

// Image represents an image for the series
type Image struct {
	CoverType string `json:"coverType"`
	URL       string `json:"url"`
	RemoteURL string `json:"remoteUrl"`
}

// Language represents the original language of the series
type Language struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

// Season represents a season in the series
type Season struct {
	SeasonNumber int  `json:"seasonNumber"`
	Monitored    bool `json:"monitored"`
}

// AddOptions represents additional options when adding the series
type AddOptions struct {
	SearchForMissingEpisodes     bool   `json:"searchForMissingEpisodes"`
	SearchForCutoffUnmetEpisodes bool   `json:"searchForCutoffUnmetEpisodes"`
	IgnoreEpisodesWithFiles      bool   `json:"ignoreEpisodesWithFiles"`
	IgnoreEpisodesWithoutFiles   bool   `json:"ignoreEpisodesWithoutFiles"`
	Monitor                      string `json:"monitor"`
}

// Ratings represents the rating details of the series
type Ratings struct {
	Votes int     `json:"votes"`
	Value float64 `json:"value"`
}

// Statistics represents statistics related to the series
type Statistics struct {
	SeasonCount       int `json:"seasonCount"`
	EpisodeFileCount  int `json:"episodeFileCount"`
	EpisodeCount      int `json:"episodeCount"`
	TotalEpisodeCount int `json:"totalEpisodeCount"`
	SizeOnDisk        int `json:"sizeOnDisk"`
	PercentOfEpisodes int `json:"percentOfEpisodes"`
}

type RequestSeriesResponse Series

type RequestSeriesBody struct {
	Title            string `json:"title"`
	TVDbId           int    `json:"tvdbId"`
	QualityProfileID int    `json:"qualityProfileId"`
	RootFolderPath   string `json:"rootFolderPath"`
	Monitored        bool   `json:"monitored"`
	AddOptions       struct {
		SearchForMissingEpisodes     bool `json:"searchForMissingEpisodes"`
		SearchForCutoffUnmetEpisodes bool `json:"searchForCutoffUnmetEpisodes"`
	}
}

type RequestShowError struct {
	PropertyName string `json:"propertyName"`
	ErrorMessage string `json:"errorMessage"`
	// AttemptedValue json.RawMessage `json:"attemptedValue"`
	Severity  string `json:"severity"`
	ErrorCode string `json:"errorCode"`
}

// Define custom error for when a show already exists in Sonarr
var ErrShowAlreadyExists = fmt.Errorf("movie already exists in sonarr")

func (r *sonarrService) RequestSeries(ctx context.Context, body RequestSeriesBody) (RequestSeriesResponse, error) {
	url, err := r.FetchSonarrURLFromDB()
	if err != nil {
		return RequestSeriesResponse{}, err
	}

	var response RequestSeriesResponse
	var responseErrors []RequestShowError

	apiResponse, err := url.New().Post("/api/v3/series").BodyJSON(body).Receive(&response, &responseErrors)
	if err != nil {
		return RequestSeriesResponse{}, errors.ErrInternalServerError().SetDetail("Failed to request series")
	}

	// Check the HTTP status code to determine if it was successful
	if apiResponse.StatusCode >= 200 && apiResponse.StatusCode < 300 {
		// Successful response, return the movie data
		return response, nil
	}

	// If Radarr returns an array of errors, iterate through them
	if len(responseErrors) > 0 {
		for _, err := range responseErrors {
			// Check if the error code is for an existing show
			if err.ErrorCode == "SeriesExistsValidator" {
				return RequestSeriesResponse{}, ErrShowAlreadyExists
			}

			// Collect other error messages
			errorMessages := fmt.Sprintf("%s - %s (Code: %s, Severity: %s)",
				err.PropertyName, err.ErrorMessage, err.ErrorCode, err.Severity)
			return RequestSeriesResponse{}, fmt.Errorf("radarr api error: %s", errorMessages)
		}
	}

	// If no specific errors were captured, return a generic error
	return RequestSeriesResponse{}, fmt.Errorf("radarr api returned an unexpected error")
}
