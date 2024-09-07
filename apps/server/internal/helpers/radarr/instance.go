package radarr

import (
	"github.com/dghubble/sling"
	"github.com/mahcks/blockbusterr/internal/global"
	"github.com/mahcks/blockbusterr/pkg/errors"
	"github.com/mahcks/blockbusterr/pkg/structures"
)

type Service interface {
	GetRootFolders() (GetRootFoldersResponse, error)
	GetQualityProfiles() (GetQualityProfilesResponse, error)
}

type radarrService struct {
	gctx global.Context
}

func (r *radarrService) FetchRadarrURLFromDB() (*sling.Sling, error) {
	// Fetch setting by key
	urlSetting, err := r.gctx.Crate().SQL.Queries().GetSettingByKey(r.gctx, structures.SettingRadarrURL.String())
	if err != nil {
		return nil, err
	}

	// If no setting is found
	if urlSetting == nil {
		return nil, errors.ErrNotFound().SetDetail("Radarr URL setting not found in the database")
	}

	// If setting is found but value is empty
	if !urlSetting.Value.Valid || urlSetting.Value.String == "" {
		return nil, errors.ErrInternalServerError().SetDetail("Radarr URL is set but empty")
	}

	apiKeySetting, err := r.gctx.Crate().SQL.Queries().GetSettingByKey(r.gctx, structures.SettingRadarrAPIKey.String())
	if err != nil {
		return nil, err
	}

	if apiKeySetting == nil {
		return nil, errors.ErrNotFound().SetDetail("Radarr API Key setting not found in the database")
	}

	if !apiKeySetting.Value.Valid || apiKeySetting.Value.String == "" {
		return nil, errors.ErrInternalServerError().SetDetail("Radarr API Key is set but empty")
	}

	base := sling.New().Base(urlSetting.Value.String).
		Set("Content-Type", "application/json").
		Set("X-Api-Key", apiKeySetting.Value.String)

	// Return the Ombi URL
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
