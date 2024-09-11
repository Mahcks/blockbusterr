package movies

import (
	"database/sql"

	"github.com/mahcks/blockbusterr/internal/db"
	"github.com/mahcks/blockbusterr/internal/global"
	"github.com/mahcks/blockbusterr/internal/helpers"
	"github.com/mahcks/blockbusterr/pkg/structures"
)

type RouteGroup struct {
	gctx    global.Context
	helpers *helpers.Helpers
}

func NewRouteGroup(gctx global.Context, helpers *helpers.Helpers) *RouteGroup {
	return &RouteGroup{
		gctx:    gctx,
		helpers: helpers,
	}
}

// Utility function to convert sql.NullInt32 to *int for JSON serialization
func nullIntToPointer(ni sql.NullInt32) *int {
	if ni.Valid {
		val := int(ni.Int32)
		return &val
	}
	return nil
}

// Utility function to convert sql.NullString to *string for JSON serialization
func nullStringToPointer(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

// Map the database allowed countries to the JSON response struct
func mapAllowedCountries(dbCountries []db.MovieAllowedCountries) []structures.MovieAllowedCountry {
	countries := make([]structures.MovieAllowedCountry, len(dbCountries))
	for i, c := range dbCountries {
		countries[i] = structures.MovieAllowedCountry{
			ID:          c.ID,
			CountryCode: c.CountryCode,
		}
	}
	return countries
}

// Map the database allowed languages to the JSON response struct
func mapAllowedLanguages(dbLanguages []db.MovieAllowedLanguages) []structures.MovieAllowedLanguage {
	languages := make([]structures.MovieAllowedLanguage, len(dbLanguages))
	for i, l := range dbLanguages {
		languages[i] = structures.MovieAllowedLanguage{
			ID:           l.ID,
			LanguageCode: l.LanguageCode,
		}
	}
	return languages
}

// Map the database blacklisted genres to the JSON response struct
func mapBlacklistedGenres(dbGenres []db.BlacklistedGenres) []structures.BlacklistedGenre {
	genres := make([]structures.BlacklistedGenre, len(dbGenres))
	for i, g := range dbGenres {
		genres[i] = structures.BlacklistedGenre{
			ID:    g.ID,
			Genre: g.Genre,
		}
	}
	return genres
}

// Map the database blacklisted title keywords to the JSON response struct
func mapBlacklistedTitleKeywords(dbKeywords []db.BlacklistedTitleKeywords) []structures.BlacklistedTitleKeyword {
	keywords := make([]structures.BlacklistedTitleKeyword, len(dbKeywords))
	for i, k := range dbKeywords {
		keywords[i] = structures.BlacklistedTitleKeyword{
			ID:      k.ID,
			Keyword: k.Keyword,
		}
	}
	return keywords
}

// Map the database blacklisted TMDB IDs to the JSON response struct
func mapBlacklistedTMDBIDs(dbTMDBIDs []db.BlacklistedTMDBIDs) []structures.BlacklistedTMDBID {
	tmdbIDs := make([]structures.BlacklistedTMDBID, len(dbTMDBIDs))
	for i, t := range dbTMDBIDs {
		tmdbIDs[i] = structures.BlacklistedTMDBID{
			ID:     t.ID,
			TMDBID: t.TMDBID,
		}
	}
	return tmdbIDs
}
