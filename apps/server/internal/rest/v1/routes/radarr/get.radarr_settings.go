package radarr

import (
	"github.com/mahcks/blockbusterr/internal/rest/v1/respond"
	"github.com/mahcks/blockbusterr/pkg/errors"
	"github.com/mahcks/blockbusterr/pkg/structures"
	"github.com/mahcks/blockbusterr/pkg/utils"
)

func (rg *RouteGroup) GetRadarrSettings(ctx *respond.Ctx) error {
	// Fetch radarr settings from the database
	settings, err := rg.gctx.Crate().SQL.Queries().GetRadarrSettings(ctx.Context())
	if err != nil {
		return errors.ErrInternalServerError().SetDetail("Failed to retrieve radarr settings")
	}

	return ctx.JSON(structures.RadarrSettings{
		ID:                  settings.ID,
		APIKey:              utils.NullStringToPointer(settings.APIKey),
		URL:                 utils.NullStringToPointer(settings.URL),
		MinimumAvailability: utils.NullStringToPointer(settings.MinimumAvailability),
		RootFolder:          utils.NullIntToPointer(settings.RootFolder),
		Quality:             utils.NullIntToPointer(settings.Quality),
	})
}
