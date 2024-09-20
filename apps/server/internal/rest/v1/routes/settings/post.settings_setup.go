package settings

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/charmbracelet/log"
	"github.com/gofiber/fiber/v2"
	"github.com/mahcks/blockbusterr/internal/rest/v1/respond"
	"github.com/mahcks/blockbusterr/pkg/errors"
)

type SetupSettings struct {
	TraktClientID        string `json:"traktClientId"`
	TraktClientSecret    string `json:"traktClientSecret"`
	OMDbAPIKey           string `json:"omdbApiKey"`
	SelectedMode         string `json:"selectedMode"`
	OmbiBaseURL          string `json:"ombi-base-url"`
	OmbiAPIKey           string `json:"ombi-api-key"`
	RadarrAPIKey         string `json:"radarr-api-key"`
	RadarrBaseURL        string `json:"radarr-base-url"`
	RadarrQualityProfile string `json:"radarr-quality-profile"`
	RadarrRootFolder     string `json:"radarr-root-folder"`
	SonarrAPIKey         string `json:"sonarr-api-key"`
	SonarrBaseURL        string `json:"sonarr-base-url"`
	SonarrQualityProfile string `json:"sonarr-quality-profile"`
	SonarrRootFolder     string `json:"sonarr-root-folder"`
}

func (rg *RouteGroup) PostSettingSetup(ctx *respond.Ctx) error {
	var payload SetupSettings
	if err := json.Unmarshal(ctx.Body(), &payload); err != nil {
		return errors.ErrBadRequest().SetDetail("Invalid JSON payload")
	}

	if payload.TraktClientID == "" {
		return errors.ErrBadRequest().SetDetail("Trakt Client ID is required")
	}

	// Create trakt settings
	err := rg.gctx.Crate().SQL.Queries().CreateTraktSettings(ctx.Context(), payload.TraktClientID, payload.TraktClientSecret)
	if err != nil {
		log.Error("error creating Trakt settings", "error", err)
		return errors.ErrInternalServerError().SetDetail("Failed to insert Trakt settings")
	}

	switch payload.SelectedMode {
	case "ombi":
		fmt.Println("OMBI SELECTED")
	case "radarr-sonarr":
		// Update radarr settings since in the database schema, there's already a row that's waiting
		convertedRadarrQualityProfile, err := strconv.Atoi(payload.RadarrQualityProfile)
		if err != nil {
			return errors.ErrBadRequest().SetDetail("Invalid Radarr quality profile")
		}

		convertedRadarrRootFolder, err := strconv.Atoi(payload.RadarrRootFolder)
		if err != nil {
			return errors.ErrBadRequest().SetDetail("Invalid Radarr root folder")
		}

		err = rg.gctx.Crate().SQL.Queries().UpdateRadarrSettings(
			ctx.Context(),
			payload.RadarrAPIKey,
			payload.RadarrBaseURL,
			"announced",
			int32(convertedRadarrQualityProfile),
			int32(convertedRadarrRootFolder),
		)
		if err != nil {
			log.Error("error updating Radarr settings", "error", err)
			return errors.ErrInternalServerError().SetDetail("Failed to update Radarr settings")
		}

		convertedSonarrQualityProfile, err := strconv.Atoi(payload.SonarrQualityProfile)
		if err != nil {
			return errors.ErrBadRequest().SetDetail("Invalid Sonarr quality profile")
		}

		convertedSonarrRootFolder, err := strconv.Atoi(payload.SonarrRootFolder)
		if err != nil {
			return errors.ErrBadRequest().SetDetail("Invalid Sonarr root folder")
		}

		// Create Sonarr settings since it's not included in the DB schema
		err = rg.gctx.Crate().SQL.Queries().CreateSonarrSettings(
			ctx.Context(),
			payload.SonarrAPIKey,
			payload.SonarrBaseURL,
			"Depracated",
			int32(convertedSonarrQualityProfile),
			int32(convertedSonarrRootFolder),
			true,
		)
		if err != nil {
			log.Error("error updating Sonarr settings", "error", err)
			return errors.ErrInternalServerError().SetDetail("Failed to update Sonarr settings")
		}

		fmt.Println("RADARR-SONARR SELECTED")
	}

	return ctx.JSON(fiber.Map{"message": "Setting inserted/updated"})
}
