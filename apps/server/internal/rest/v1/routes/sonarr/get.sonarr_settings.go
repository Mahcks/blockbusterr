package sonarr

import (
	"github.com/mahcks/blockbusterr/internal/rest/v1/respond"
	"github.com/mahcks/blockbusterr/pkg/errors"
	"github.com/mahcks/blockbusterr/pkg/structures"
	"github.com/mahcks/blockbusterr/pkg/utils"
)

func (rg *RouteGroup) GetSonarrSettings(ctx *respond.Ctx) error {
	settings, err := rg.gctx.Crate().SQL.Queries().GetSonarrSettings(ctx.Context())
	if err != nil {
		return errors.ErrInternalServerError().SetDetail("Failed to retrieve sonarr settings")
	}

	return ctx.JSON(structures.SonarrSettings{
		ID:           settings.ID,
		APIKey:       utils.NullStringToPointer(settings.APIKey),
		URL:          utils.NullStringToPointer(settings.URL),
		Language:     utils.NullStringToPointer(settings.Language),
		Quality:      utils.NullIntToPointer(settings.Quality),
		RootFolder:   utils.NullIntToPointer(settings.RootFolder),
		SeasonFolder: utils.NullBoolToPointer(settings.SeasonFolder),
	})
}
