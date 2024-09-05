package settings

import (
	"encoding/json"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/mahcks/blockbusterr/internal/rest/v1/respond"
	"github.com/mahcks/blockbusterr/pkg/errors"
	"github.com/mahcks/blockbusterr/pkg/structures"
)

// UpdateSetting handles the updating of a setting
func (rg *RouteGroup) PutSetting(ctx *respond.Ctx) error {
	// Get raw request body
	body := ctx.Body()

	// Initialize payload to store unmarshalled JSON
	var payload SettingPayload

	// Unmarshal raw JSON body into the struct
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Printf("Failed to unmarshal JSON payload: %v", err)
		return errors.ErrBadRequest().SetDetail("Failed to parse JSON payload")
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

	// Update the setting in the database
	err := rg.gctx.Crate().SQL.Queries().InsertOrUpdateSetting(ctx.Context(), payload.Key, payload.Value, payload.Type)
	if err != nil {
		return errors.ErrInternalServerError().SetDetail("Failed to update setting")
	}

	return ctx.JSON(fiber.Map{"message": "Setting updated"})
}
