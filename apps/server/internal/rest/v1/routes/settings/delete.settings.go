package settings

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/mahcks/blockbusterr/internal/rest/v1/respond"
	"github.com/mahcks/blockbusterr/pkg/errors"
	"github.com/mahcks/blockbusterr/pkg/structures"
)

// DeleteSetting handles the deletion of a setting by its key
func (rg *RouteGroup) DeleteSetting(ctx *respond.Ctx) error {
	// Get the key parameter from the query string
	key := ctx.Query("key")
	if key == "" {
		return errors.ErrBadRequest().SetDetail("Key is required")
	}

	// Validate the setting key against known constants
	if !structures.IsValidSettingKey(structures.Setting(key)) {
		return errors.ErrBadRequest().SetDetail("Invalid setting key provided")
	}

	// Delete the setting from the database
	err := rg.gctx.Crate().SQL.Queries().DeleteSettingByKey(context.Background(), key)
	if err != nil {
		log.Printf("Failed to delete setting: %v", err)
		return errors.ErrInternalServerError().SetDetail("Failed to delete setting")
	}

	return ctx.JSON(fiber.Map{
		"message": "Setting deleted",
	})
}
