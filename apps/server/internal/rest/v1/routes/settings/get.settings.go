package settings

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/mahcks/blockbusterr/internal/rest/v1/respond"
	"github.com/mahcks/blockbusterr/pkg/errors"
	"github.com/mahcks/blockbusterr/pkg/structures"
)

// GetSetting handles retrieving a setting by its key
func (rg *RouteGroup) GetSetting(ctx *respond.Ctx) error {
	// Get the key parameter from the query string
	key := ctx.Query("key")
	if key == "" {
		return errors.ErrBadRequest().SetDetail("Key is required")
	}

	// Validate the setting key against known constants
	if !structures.IsValidSettingKey(structures.Setting(key)) {
		return errors.ErrBadRequest().SetDetail("Invalid setting key provided")
	}

	// Fetch the setting from the database
	setting, err := rg.gctx.Crate().SQL.Queries().GetSettingByKey(context.Background(), key)
	if err != nil {
		log.Printf("Failed to get setting from DB: %v", err)
		return errors.ErrInternalServerError().SetDetail("Failed to retrieve setting")
	}

	// Return the setting value
	return ctx.JSON(fiber.Map{
		"key":   setting.Key,
		"value": setting.Value.String,
		"type":  setting.Type.String,
	})
}
