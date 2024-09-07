package ombi

import (
	"github.com/dghubble/sling"
	"github.com/mahcks/blockbusterr/internal/global"
	"github.com/mahcks/blockbusterr/pkg/errors"
	"github.com/mahcks/blockbusterr/pkg/structures"
)

type Service interface {
	// Radarr

	GetRadarrEnabled() (bool, error)
}

type ombiService struct {
	gctx global.Context
}

func (o *ombiService) FetchOmbiURLFromDB() (*sling.Sling, error) {
	// Fetch setting by key
	urlSetting, err := o.gctx.Crate().SQL.Queries().GetSettingByKey(o.gctx, structures.SettingOmbiURL.String())
	if err != nil {
		return nil, err
	}

	// If no setting is found
	if urlSetting == nil {
		return nil, errors.ErrNotFound().SetDetail("Ombi URL setting not found in the database")
	}

	// If setting is found but value is empty
	if !urlSetting.Value.Valid || urlSetting.Value.String == "" {
		return nil, errors.ErrInternalServerError().SetDetail("Ombi URL is set but empty")
	}

	apiKeySetting, err := o.gctx.Crate().SQL.Queries().GetSettingByKey(o.gctx, structures.SettingOmbiAPIKey.String())
	if err != nil {
		return nil, err
	}

	if apiKeySetting == nil {
		return nil, errors.ErrNotFound().SetDetail("Ombi API Key setting not found in the database")
	}

	if !apiKeySetting.Value.Valid || apiKeySetting.Value.String == "" {
		return nil, errors.ErrInternalServerError().SetDetail("Ombi API Key is set but empty")
	}

	base := sling.New().Base(urlSetting.Value.String).
		Set("Content-Type", "application/json").
		Set("ApiKey", apiKeySetting.Value.String)

	// Return the Ombi URL
	return base, nil
}

func (o *ombiService) GetRadarrEnabled() (bool, error) {
	url, err := o.FetchOmbiURLFromDB()
	if err != nil {
		return false, err
	}

	var isEnabled bool
	_, err = url.QueryStruct(nil).Get("/api/v1/radarr/enabled").ReceiveSuccess(&isEnabled)
	if err != nil {
		return false, errors.ErrInternalServerError().SetDetail("Failed to get Radarr enabled status from Ombi")
	}

	return isEnabled, nil
}
