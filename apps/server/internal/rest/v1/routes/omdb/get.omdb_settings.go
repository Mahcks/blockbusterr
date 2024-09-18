package omdb

import (
	"github.com/mahcks/blockbusterr/internal/rest/v1/respond"
	"github.com/mahcks/blockbusterr/pkg/errors"
	"github.com/mahcks/blockbusterr/pkg/structures"
	"github.com/mahcks/blockbusterr/pkg/utils"
)

func (rg *RouteGroup) GetOMDbSettings(ctx *respond.Ctx) error {
	// Fetch omdb settings from the database
	settings, err := rg.gctx.Crate().SQL.Queries().GetOMDbSettings(ctx.Context())
	if err != nil {
		return errors.ErrInternalServerError().SetDetail("Failed to retrieve trakt settings")
	}

	// Map the DB TraktSettings struct to the structures.TraktSettings struct for JSON response
	response := structures.OMDbSettings{
		ID:     settings.ID,
		APIKey: utils.NullStringToPointer(settings.APIKey),
	}

	return ctx.JSON(response)
}
