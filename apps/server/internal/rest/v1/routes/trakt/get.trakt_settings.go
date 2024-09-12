package trakt

import (
	"github.com/mahcks/blockbusterr/internal/rest/v1/respond"
	"github.com/mahcks/blockbusterr/pkg/errors"
	"github.com/mahcks/blockbusterr/pkg/structures"
)

func (rg *RouteGroup) GetTraktSettings(ctx *respond.Ctx) error {
	// Fetch trakt settings from the database
	settings, err := rg.gctx.Crate().SQL.Queries().GetTraktSettings(ctx.Context())
	if err != nil {
		return errors.ErrInternalServerError().SetDetail("Failed to retrieve trakt settings")
	}

	// Map the DB TraktSettings struct to the structures.TraktSettings struct for JSON response
	response := structures.TraktSettings{
		ID:           settings.ID,
		ClientID:     settings.ClientID,
		ClientSecret: settings.ClientSecret,
	}

	return ctx.JSON(response)
}
