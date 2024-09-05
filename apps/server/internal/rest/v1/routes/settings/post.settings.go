package settings

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/mahcks/blockbusterr/internal/rest/v1/respond"
	"github.com/mahcks/blockbusterr/pkg/errors"
	"github.com/mahcks/blockbusterr/pkg/structures"
)

func (rg *RouteGroup) PostSetting(ctx *respond.Ctx) error {
	// Parse JSON body into SettingPayload struct
	var payload SettingPayload
	if err := json.Unmarshal(ctx.Body(), &payload); err != nil {
		return errors.ErrBadRequest().SetDetail("Invalid JSON payload")
	}

	// Validate the setting key against known constants
	if !structures.IsValidSettingKey(structures.Setting(payload.Key)) {
		return errors.ErrBadRequest().SetDetail("Invalid setting key provided")
	}

	// Validate required fields
	if payload.Value == "" {
		return errors.ErrBadRequest().SetDetail("Key and value are required")
	}

	// Default the type to "text" if it's not provided
	if payload.Type == "" {
		payload.Type = "text"
	}

	// Insert or update the setting in the database
	err := rg.gctx.Crate().SQL.Queries().InsertOrUpdateSetting(ctx.Context(), payload.Key, payload.Value, payload.Type)
	if err != nil {
		return errors.ErrInternalServerError().SetDetail("Failed to insert/update setting")
	}

	return ctx.JSON(fiber.Map{"message": "Setting inserted/updated"})
}
