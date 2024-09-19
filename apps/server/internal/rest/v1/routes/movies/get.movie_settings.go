package movies

import (
	"github.com/mahcks/blockbusterr/internal/rest/v1/respond"
	"github.com/mahcks/blockbusterr/pkg/errors"
	"github.com/mahcks/blockbusterr/pkg/structures"
)

func (rg *RouteGroup) GetMovieSettings(ctx *respond.Ctx) error {
	// Fetch movie settings from database
	settings, err := rg.gctx.Crate().SQL.Queries().GetMovieSettings(ctx.Context())
	if err != nil {
		return errors.ErrInternalServerError().SetDetail("Failed to retrieve movie settings")
	}

	// Map the DB MovieSettings struct to the structures.MovieSettings struct for JSON response
	response := structures.MovieSettings{
		ID:                       settings.ID,
		Anticipated:              nullIntToPointer(settings.Anticipated),
		CronJobAnticipated:       nullStringToPointer(settings.CronAnticipated),
		BoxOffice:                nullIntToPointer(settings.BoxOffice),
		CronJobBoxOffice:         nullStringToPointer(settings.CronBoxOffice),
		Popular:                  nullIntToPointer(settings.Popular),
		CronJobPopular:           nullStringToPointer(settings.CronPopular),
		Trending:                 nullIntToPointer(settings.Trending),
		CronJobTrending:          nullStringToPointer(settings.CronTrending),
		MaxRuntime:               nullIntToPointer(settings.MaxRuntime),
		MinRuntime:               nullIntToPointer(settings.MinRuntime),
		MinYear:                  nullIntToPointer(settings.MinYear),
		MaxYear:                  nullIntToPointer(settings.MaxYear),
		RottenTomatoes:           nullStringToPointer(settings.RottenTomatoes),
		AllowedCountries:         mapAllowedCountries(settings.AllowedCountries),
		AllowedLanguages:         mapAllowedLanguages(settings.AllowedLanguages),
		BlacklistedGenres:        mapBlacklistedGenres(settings.BlacklistedGenres),
		BlacklistedTitleKeywords: mapBlacklistedTitleKeywords(settings.BlacklistedTitleKeywords),
		BlacklistedTMDBIDs:       mapBlacklistedTMDBIDs(settings.BlacklistedTMDBIDs),
	}

	return ctx.JSON(response)
}
