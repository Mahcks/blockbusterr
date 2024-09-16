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
		Interval:                 nullIntToPointer(settings.Interval), // Convert sql.NullInt32 to *int
		Anticipated:              nullIntToPointer(settings.Anticipated),
		BoxOffice:                nullIntToPointer(settings.BoxOffice),
		Popular:                  nullIntToPointer(settings.Popular),
		Trending:                 nullIntToPointer(settings.Trending),
		MaxRuntime:               nullIntToPointer(settings.MaxRuntime),
		MinRuntime:               nullIntToPointer(settings.MinRuntime),
		MinYear:                  nullIntToPointer(settings.MinYear),
		MaxYear:                  nullIntToPointer(settings.MaxYear),
		RottenTomatoes:           nullStringToPointer(settings.RottenTomatoes), // Convert sql.NullString to *string
		AllowedCountries:         mapAllowedCountries(settings.AllowedCountries),
		AllowedLanguages:         mapAllowedLanguages(settings.AllowedLanguages),
		BlacklistedGenres:        mapBlacklistedGenres(settings.BlacklistedGenres),
		BlacklistedTitleKeywords: mapBlacklistedTitleKeywords(settings.BlacklistedTitleKeywords),
		BlacklistedTMDBIDs:       mapBlacklistedTMDBIDs(settings.BlacklistedTMDBIDs),
	}

	return ctx.JSON(response)
}
