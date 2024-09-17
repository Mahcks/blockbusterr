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
		Anticipated:              utils.NullIntToPointer(settings.Anticipated),
		CronJobAnticipated:       utils.NullStringToPointer(settings.CronJobAnticipated),
		Popular:                  utils.NullIntToPointer(settings.Popular),
		CronJobPopular:           utils.NullStringToPointer(settings.CronJobPopular),
		Trending:                 utils.NullIntToPointer(settings.Trending),
		CronJobTrending:          utils.NullStringToPointer(settings.CronJobTrending),
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
