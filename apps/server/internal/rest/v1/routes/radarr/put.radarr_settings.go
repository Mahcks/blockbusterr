package radarr

import (
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/mahcks/blockbusterr/internal/rest/v1/respond"
	"github.com/mahcks/blockbusterr/pkg/errors"
	"github.com/mahcks/blockbusterr/pkg/utils"
)

type RadarrSettingsPayload struct {
	APIKey              *string `json:"api_key"`
	URL                 *string `json:"base_url"`
	MinimumAvailability *string `json:"minimum_availability"`
	QualityProfile      *int    `json:"quality_profile"`
	RootFolder          *int    `json:"root_folder"`
}

func (rg *RouteGroup) UpdateRadarrSettings(ctx *respond.Ctx) error {
	var payload RadarrSettingsPayload
	if err := json.Unmarshal(ctx.Body(), &payload); err != nil {
		fmt.Println(err)
		return errors.ErrBadRequest().SetDetail("Invalid JSON payload")
	}

	utils.PrettyPrintStruct(payload)
	err := rg.gctx.Crate().SQL.Queries().UpdateRadarrSettings(
		ctx.Context(),
		utils.PointerToNullString(payload.APIKey),
		utils.PointerToNullString(payload.URL),
		utils.PointerToNullString(payload.MinimumAvailability),
		utils.PointerToNullInt32(payload.QualityProfile),
		utils.PointerToNullInt32(payload.RootFolder),
	)
	if err != nil {
		return errors.ErrInternalServerError().SetDetail("Failed to update Radarr settings")
	}

	return ctx.JSON(fiber.Map{"success": true})
}
