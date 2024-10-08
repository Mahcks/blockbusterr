package ombi

import (
	"fmt"

	"github.com/dghubble/sling"
	"github.com/mahcks/blockbusterr/internal/global"
	"github.com/mahcks/blockbusterr/pkg/errors"
)

type Service interface {
	// Ombi

	GetUsers() (GetUsersResponse, error)

	// Radarr

	GetRadarrEnabled() (bool, error)
	GetRadarrProfiles() (GetRadarrProfilesResponse, error)
	GetRadarrRootFolders() (GetRadarrRootFoldersResponse, error)
	RequestMovie(body RequestMovieBody) (RequestMovieResponse, error)

	// Sonarr
	GetSonarrEnabled() (bool, error)
	GetSonnarProfiles() (GetSonarrProfilesResponse, error)
	GetSonarrRootFolders() (GetSonarrRootFoldersResponse, error)
	RequestShow(body RequestShowBody) (RequestShowResponse, error)
}

type ombiService struct {
	gctx global.Context
}

func (o *ombiService) FetchOmbiURLFromDB() (*sling.Sling, error) {
	// Get Ombi enabled setting
	ombiSettings, err := o.gctx.Crate().SQL.Queries().GetOmbiSettings(o.gctx)
	if err != nil {
		return nil, err
	}

	// If setting is found but value is empty
	if !ombiSettings.URL.Valid || ombiSettings.URL.String == "" {
		return nil, errors.ErrInternalServerError().SetDetail("Ombi URL is set but empty")
	}

	if !ombiSettings.APIKey.Valid || ombiSettings.APIKey.String == "" {
		return nil, errors.ErrInternalServerError().SetDetail("Ombi API Key is set but empty")
	}

	base := sling.New().Base(ombiSettings.URL.String).
		Set("Content-Type", "application/json").
		Set("ApiKey", ombiSettings.APIKey.String)

	// Return the Ombi URL
	return base, nil
}

type User struct {
	ID                      string             `json:"id"`
	UserName                string             `json:"userName"`
	Alias                   string             `json:"alias,omitempty"`
	Claims                  []UserClaims       `json:"claims"`
	EmailAddress            string             `json:"emailAddress"`
	LastLoggedIn            string             `json:"lastLoggedIn"`
	HasLoggedIn             bool               `json:"hasLoggedIn"`
	UserType                int                `json:"userType"`
	MovieRequestLimit       int                `json:"movieRequestLimit"`
	EpisodeRequestLimit     int                `json:"episodeRequestLimit"`
	StreamingCountry        string             `json:"streamingCountry"`
	EpisodeRequestQuota     int                `json:"episodeRequestQuota,omitempty"`
	MovieRequestQuota       int                `json:"movieRequestQuota,omitempty"`
	UserQualityProfiles     UserQualityProfile `json:"userQualityProfiles"`
	MovieRequestLimitType   int                `json:"movieRequestLimitType"`
	MusicRequestLimitType   int                `json:"musicRequestLimitType"`
	EpisodeRequestLimitType int                `json:"episodeRequestLimitType"`
}

type UserClaims struct {
	Value   string `json:"value"`
	Enabled bool   `json:"enabled"`
}

type UserQualityProfile struct {
	UserID                    string `json:"userId"`
	SonarrQualityProfileAnime int    `json:"sonarrQualityProfileAnime"`
	SonarrRootPathAnime       int    `json:"sonarrRootPathAnime"`
	SonarrQualityProfile      int    `json:"sonarrQualityProfile"`
	RadarrRootPath            int    `json:"radarrRootPath"`
	RadarrQualityProfile      int    `json:"radarrQualityProfile"`
	Radarr4KRootPath          int    `json:"radarr4KRootPath"`
	Radarr4KQualityProfile    int    `json:"radarr4KQualityProfile"`
	ID                        int    `json:"id"`
}

type GetUsersResponse []User

func (o *ombiService) GetUsers() (GetUsersResponse, error) {
	url, err := o.FetchOmbiURLFromDB()
	if err != nil {
		return nil, err
	}

	var users GetUsersResponse
	_, err = url.Get("/api/v1/identity/users").ReceiveSuccess(&users)
	if err != nil {
		return nil, errors.ErrInternalServerError().SetDetail("Failed to get Ombi users")
	}

	return users, nil
}

func (o *ombiService) GetRadarrEnabled() (bool, error) {
	url, err := o.FetchOmbiURLFromDB()
	if err != nil {
		return false, err
	}

	var isEnabled bool
	_, err = url.Get("/api/v1/radarr/enabled").ReceiveSuccess(&isEnabled)
	if err != nil {
		return false, errors.ErrInternalServerError().SetDetail("Failed to get Radarr enabled status from Ombi")
	}

	return isEnabled, nil
}

type GetRadarrProfilesResponse []QualityProfile

type QualityProfile struct {
	ID             int                  `json:"id"`
	Name           string               `json:"name"`
	UpgradeAllowed bool                 `json:"upgradeAllowed"`
	Cutoff         int                  `json:"cutoff"`
	PreferredTags  []string             `json:"preferredTags,omitempty"`
	Items          []QualityProfileItem `json:"items"`
}

type QualityProfileItem struct {
	Quality *Quality             `json:"quality,omitempty"` // Pointer to omit empty values
	Items   []QualityProfileItem `json:"items,omitempty"`   // Nested items
	Allowed bool                 `json:"allowed"`
}

type Quality struct {
	ID         int    `json:"id,omitempty"`
	Name       string `json:"name,omitempty"`
	Source     string `json:"source,omitempty"`
	Resolution int    `json:"resolution,omitempty"`
	Modifier   string `json:"modifier,omitempty"`
}

func (o *ombiService) GetRadarrProfiles() (GetRadarrProfilesResponse, error) {
	url, err := o.FetchOmbiURLFromDB()
	if err != nil {
		return nil, err
	}

	var response GetRadarrProfilesResponse
	_, err = url.Get("/api/v1/radarr/Profiles").ReceiveSuccess(&response)
	if err != nil {
		return nil, errors.ErrInternalServerError().SetDetail("Failed to get Radarr quality profiles via Ombi")
	}

	return response, nil
}

type GetRadarrRootFoldersResponse []RadarrRootFolder

type RadarrRootFolder struct {
	ID        int    `json:"id"`
	Path      string `json:"path"`
	Freespace int64  `json:"freespace"`
}

func (o *ombiService) GetRadarrRootFolders() (GetRadarrRootFoldersResponse, error) {
	url, err := o.FetchOmbiURLFromDB()
	if err != nil {
		return nil, err
	}

	var response GetRadarrRootFoldersResponse
	_, err = url.Get("/api/v1/radarr/rootfolders").ReceiveSuccess(&response)
	if err != nil {
		return nil, errors.ErrInternalServerError().SetDetail("Failed to get Radarr quality profiles via Ombi")
	}

	return response, nil
}

type RequestMovieBody struct {
	TheMovieDBID        int     `json:"theMovieDbId"`
	LanguageCode        string  `json:"languageCode"`
	QualityPathOverride *string `json:"qualityProfileOverride"`
	RequestOnBehalf     string  `json:"requestOnBehalf"`
	RootFolderOverride  *string `json:"rootFolderOverride"`
	Is4KRequest         bool    `json:"is4kRequest"`
}

type RequestMovieResponse struct {
	Result       bool   `json:"result"`
	Message      string `json:"message"`
	IsError      bool   `json:"isError"`
	ErrorMessage string `json:"errorMessage"`
	ErrorCode    string `json:"errorCode"`
	RequestID    int    `json:"requestId"`
}

var ErrMovieAlreadyRequested = fmt.Errorf("movie already requested")

func (o *ombiService) RequestMovie(body RequestMovieBody) (RequestMovieResponse, error) {
	url, err := o.FetchOmbiURLFromDB()
	if err != nil {
		return RequestMovieResponse{}, err
	}

	var response RequestMovieResponse
	_, err = url.New().Post("/api/v1/request/movie").BodyJSON(body).ReceiveSuccess(&response)
	if err != nil {
		return RequestMovieResponse{}, errors.ErrInternalServerError().SetDetail("Failed to request movie via Ombi")
	}

	if response.ErrorCode == "MovieAlreadyRequested" || response.ErrorCode == "AlreadyRequested" {
		return response, ErrMovieAlreadyRequested
	}

	return response, nil
}

// SONARR

func (o *ombiService) GetSonarrEnabled() (bool, error) {
	url, err := o.FetchOmbiURLFromDB()
	if err != nil {
		return false, err
	}

	var isEnabled bool
	_, err = url.QueryStruct(nil).Get("/api/v1/sonarr/enabled").ReceiveSuccess(&isEnabled)
	if err != nil {
		return false, errors.ErrInternalServerError().SetDetail("Failed to get Sonarr enabled status from Ombi")
	}

	return isEnabled, nil
}

type GetSonarrProfilesResponse []SonarrQualityProfile

type SonarrQualityProfile struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (o *ombiService) GetSonnarProfiles() (GetSonarrProfilesResponse, error) {
	url, err := o.FetchOmbiURLFromDB()
	if err != nil {
		return nil, err
	}

	var response GetSonarrProfilesResponse
	_, err = url.Get("/api/v1/sonarr/profiles").ReceiveSuccess(&response)
	if err != nil {
		return nil, errors.ErrInternalServerError().SetDetail("Failed to get Sonarr quality profiles via Ombi")
	}

	return response, nil
}

type GetSonarrRootFoldersResponse []SonarrRootFolder

type SonarrRootFolder struct {
	ID        int    `json:"id"`
	Path      string `json:"path"`
	Freespace int64  `json:"freespace"`
}

func (o *ombiService) GetSonarrRootFolders() (GetSonarrRootFoldersResponse, error) {
	url, err := o.FetchOmbiURLFromDB()
	if err != nil {
		return nil, err
	}

	var response GetSonarrRootFoldersResponse
	_, err = url.Get("/api/v1/sonarr/rootfolders").ReceiveSuccess(&response)
	if err != nil {
		return nil, errors.ErrInternalServerError().SetDetail("Failed to get Sonarr root folders via Ombi")
	}

	return response, nil
}

type RequestShowBody struct {
	FirstSeason         bool    `json:"firstSeason"`
	LatestSeason        bool    `json:"latestSeason"`
	RequestAll          bool    `json:"requestAll"`
	TheMovieDBID        int     `json:"theMovieDbId"`
	RequestOnBehalf     string  `json:"requestOnBehalf"`
	LanguageCode        string  `json:"languageCode"`
	QualityPathOverride *string `json:"qualityProfileOverride"`
	RootFolderOverride  *string `json:"rootFolderOverride"`
}

type RequestShowResponse struct {
	Result       bool   `json:"result"`
	Message      string `json:"message"`
	IsError      bool   `json:"isError"`
	ErrorMessage string `json:"errorMessage"`
	ErrorCode    string `json:"errorCode"`
	RequestID    int    `json:"requestId"`
}

var ErrShowAlreadyRequested = fmt.Errorf("show already requested")

func (o *ombiService) RequestShow(body RequestShowBody) (RequestShowResponse, error) {
	url, err := o.FetchOmbiURLFromDB()
	if err != nil {
		return RequestShowResponse{}, err
	}

	var response RequestShowResponse
	_, err = url.New().Post("/api/v2/requests/tv").BodyJSON(body).ReceiveSuccess(&response)
	if err != nil {
		return RequestShowResponse{}, errors.ErrInternalServerError().SetDetail("Failed to request movie via Ombi")
	}

	if response.ErrorCode == "MovieAlreadyRequested" || response.ErrorCode == "AlreadyRequested" || response.IsError {
		return response, ErrShowAlreadyRequested
	}

	return RequestShowResponse(response), nil
}
