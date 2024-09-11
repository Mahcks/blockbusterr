package shows

import (
	"github.com/mahcks/blockbusterr/internal/rest/v1/respond"
	"github.com/mahcks/blockbusterr/pkg/errors"
	"github.com/mahcks/blockbusterr/pkg/structures"
	"github.com/mahcks/blockbusterr/pkg/utils"
)

func (rg *RouteGroup) GetShowSettings(ctx *respond.Ctx) error {
	// Fetch show settings from the database
	settings, err := rg.gctx.Crate().SQL.Queries().GetShowSettings(ctx.Context())
	if err != nil {
		return errors.ErrInternalServerError().SetDetail("Failed to retrieve show settings")
	}

	// Map the DB ShowSettings struct to the structures.ShowSettings struct for JSON response
	response := structures.ShowSettings{
		ID:                       settings.ID,
		Interval:                 utils.NullIntToPointer(settings.Interval), // Convert sql.NullInt32 to *int
		Anticipated:              utils.NullIntToPointer(settings.Anticipated),
		Popular:                  utils.NullIntToPointer(settings.Popular),
		Trending:                 utils.NullIntToPointer(settings.Trending),
		MaxRuntime:               utils.NullIntToPointer(settings.MaxRuntime),
		MinRuntime:               utils.NullIntToPointer(settings.MinRuntime),
		MinYear:                  utils.NullIntToPointer(settings.MinYear),
		MaxYear:                  utils.NullIntToPointer(settings.MaxYear),
		AllowedCountries:         mapAllowedCountries(settings.AllowedCountries),
		AllowedLanguages:         mapAllowedLanguages(settings.AllowedLanguages),
		BlacklistedGenres:        mapBlacklistedGenres(settings.BlacklistedGenres),
		BlacklistedNetworks:      mapBlacklistedNetworks(settings.BlacklistedNetworks),
		BlacklistedTitleKeywords: mapBlacklistedShowTitleKeywords(settings.BlacklistedTitleKeywords),
		BlacklistedTVDBIDs:       mapBlacklistedTVDBIDs(settings.BlacklistedTVDBIDs),
	}

	return ctx.JSON(response)
}
