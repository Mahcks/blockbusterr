package radarr

import (
	"time"

	"github.com/dghubble/sling"
	"github.com/mahcks/blockbusterr/internal/global"
	"github.com/mahcks/blockbusterr/pkg/errors"
)

type Service interface {
	GetRootFolders() (GetRootFoldersResponse, error)
	GetQualityProfiles() (GetQualityProfilesResponse, error)
	RequestMovie(RequestMovieBody) (RequestMovieResponse, error)
}

type radarrService struct {
	gctx global.Context
}

func (r *radarrService) FetchRadarrURLFromDB() (*sling.Sling, error) {
	radarrSettings, err := r.gctx.Crate().SQL.Queries().GetRadarrSettings(r.gctx)
	if err != nil {
		return nil, err
	}

	// Check if Radarr URL is valid and not empty
	if !radarrSettings.URL.Valid || radarrSettings.URL.String == "" {
		return nil, errors.ErrInternalServerError().SetDetail("Radarr URL is set but empty")
	}

	// Check if Radarr API Key is valid and not empty
	if !radarrSettings.APIKey.Valid || radarrSettings.APIKey.String == "" {
		return nil, errors.ErrInternalServerError().SetDetail("Radarr API Key is set but empty")
	}

	base := sling.New().Base(radarrSettings.URL.String).
		Set("Content-Type", "application/json").
		Set("X-Api-Key", radarrSettings.APIKey.String)

	// Return the configured Radarr URL
	return base, nil
}

type GetRootFoldersResponse []RootFolder

type RootFolder struct {
	ID              int                        `json:"id"`
	Path            string                     `json:"path"`
	Accessible      bool                       `json:"accessible"`
	FreeSpace       int                        `json:"freeSpace"`
	UnmappedFolders []RootFolderUnmappedFolder `json:"unmappedFolders"`
}

type RootFolderUnmappedFolder struct {
	Name         string `json:"name"`
	Path         string `json:"path"`
	RelativePath string `json:"relativePath"`
}

func (r *radarrService) GetRootFolders() (GetRootFoldersResponse, error) {
	url, err := r.FetchRadarrURLFromDB()
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

type QualityProfile struct {
	ID                int                    `json:"id"`
	Name              string                 `json:"name"`
	UpgradeAllowed    bool                   `json:"upgradeAllowed"`
	Cutoff            int                    `json:"cutoff"`
	MinFormatScore    int                    `json:"minFormatScore"`
	CutoffFormatScore int                    `json:"cutoffFormatScore"`
	Items             QualityProfileItem     `json:"items"`
	Language          QualityProfileLanguage `json:"language"`
}

type QualityProfileItem []QualityProfileItemData

type QualityProfileItemData struct {
	Quality struct {
		ID         int    `json:"id"`
		Name       string `json:"name"`
		Source     string `json:"source"`
		Resolution int    `json:"resolution"`
		Modifier   string `json:"modifier"`
	} `json:"quality"`
	Allowed bool `json:"allowed"`
}

type QualityProfileLanguage struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (r *radarrService) GetQualityProfiles() (GetQualityProfilesResponse, error) {
	url, err := r.FetchRadarrURLFromDB()
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

type Movie struct {
	Title                 string                     `json:"title"`
	OriginalTitle         string                     `json:"originalTitle"`
	OriginalLanguage      Language                   `json:"originalLanguage"`
	AlternateTitles       []AlternativeTitleResource `json:"alternateTitles"`
	SecondaryYear         *int                       `json:"secondaryYear"`
	SecondaryYearSourceId int                        `json:"secondaryYearSourceId"`
	SortTitle             string                     `json:"sortTitle"`
	SizeOnDisk            *int64                     `json:"sizeOnDisk"`
	Status                string                     `json:"status"`
	Overview              string                     `json:"overview"`
	InCinemas             *time.Time                 `json:"inCinemas"`
	PhysicalRelease       *time.Time                 `json:"physicalRelease"`
	DigitalRelease        *time.Time                 `json:"digitalRelease"`
	ReleaseDate           *time.Time                 `json:"releaseDate"`
	PhysicalReleaseNote   string                     `json:"physicalReleaseNote"`
	Images                []MediaCover               `json:"images"`
	Website               string                     `json:"website"`
	RemotePoster          string                     `json:"remotePoster"`
	Year                  int                        `json:"year"`
	YouTubeTrailerId      string                     `json:"youTubeTrailerId"`
	Studio                string                     `json:"studio"`
	Path                  string                     `json:"path"`
	QualityProfileId      int                        `json:"qualityProfileId"`
	HasFile               *bool                      `json:"hasFile"`
	MovieFileId           int                        `json:"movieFileId"`
	Monitored             bool                       `json:"monitored"`
	MinimumAvailability   string                     `json:"minimumAvailability"`
	IsAvailable           bool                       `json:"isAvailable"`
	FolderName            string                     `json:"folderName"`
	Runtime               int                        `json:"runtime"`
	CleanTitle            string                     `json:"cleanTitle"`
	ImdbId                string                     `json:"imdbId"`
	TmdbId                int                        `json:"tmdbId"`
	TitleSlug             string                     `json:"titleSlug"`
	RootFolderPath        string                     `json:"rootFolderPath"`
	Folder                string                     `json:"folder"`
	Certification         string                     `json:"certification"`
	Genres                []string                   `json:"genres"`
	Tags                  map[int]struct{}           `json:"tags"` // Use a map to simulate HashSet
	Added                 time.Time                  `json:"added"`
	Ratings               Ratings                    `json:"ratings"`
	Popularity            float32                    `json:"popularity"`
	Statistics            MovieStatisticsResource    `json:"statistics"`
}

type Language struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type AlternativeTitleResource struct {
	Title    string `json:"title"`
	Type     string `json:"type"`
	Language string `json:"language"`
}

type MediaCover struct {
	CoverType string `json:"coverType"`
	URL       string `json:"url"`
	RemoteURL string `json:"remoteUrl"`
}

// Ratings holds various rating sources like IMDb, TMDb, Metacritic, and Rotten Tomatoes.
type Ratings struct {
	Imdb           RatingChild `json:"imdb"`
	Tmdb           RatingChild `json:"tmdb"`
	Metacritic     RatingChild `json:"metacritic"`
	RottenTomatoes RatingChild `json:"rottenTomatoes"`
}

// RatingChild holds information about votes, value, and the type of rating (User or Critic).
type RatingChild struct {
	Votes int     `json:"votes"`
	Value float64 `json:"value"` // Use float64 instead of decimal
	Type  string  `json:"type"`
}

type MovieStatisticsResource struct {
	MovieFileCount int      `json:"movieFileCount"`
	SizeOnDisk     int64    `json:"sizeOnDisk"`
	ReleaseGroups  []string `json:"releaseGroups"`
}

type RequestMovieResponse Movie

type RequestMovieBody struct {
	// Title of the film (provided by Trakt)
	Title string `json:"title"`
	// The quality profile ID to use for the film (provided by Radarr)
	QualityProfileID int `json:"qualityProfileId"`
	// The root folder to use for the film. The user can choose from the list of available root folders provided by Radarr in the settings.
	RootFolderPath string `json:"path"`
	// The Movie Database ID of the film (provided by Trakt)
	TMDBID int `json:"tmdbId"`
	// The minimum availability setting for the film. The user can choose from "announced", "in_cinemas", or "released".
	MinimumAvailability string `json:"minimumAvailability"`
	AddOptions          struct {
		// Whether to search for the film when added to Radarr
		SearchForMovie bool `json:"searchForMovie"`
	} `json:"addOptions"`
}

func (r *radarrService) RequestMovie(RequestMovieBody) (RequestMovieResponse, error) {
	url, err := r.FetchRadarrURLFromDB()
	if err != nil {
		return RequestMovieResponse{}, err
	}

	var response RequestMovieResponse
	_, err = url.New().Get("/api/v3/qualityprofile").Receive(&response, nil)
	if err != nil {
		return RequestMovieResponse{}, errors.ErrInternalServerError().SetDetail("Failed to get Radarr quality profiles")
	}

	return response, nil
}
