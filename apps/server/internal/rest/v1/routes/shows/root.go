package shows

import (
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

func mapAllowedCountries(dbCountries []db.ShowAllowedCountries) []structures.ShowAllowedCountry {
	countries := make([]structures.ShowAllowedCountry, len(dbCountries))
	for i, c := range dbCountries {
		countries[i] = structures.ShowAllowedCountry{
			ID:          c.ID,
			CountryCode: c.CountryCode,
		}
	}
	return countries
}

func mapAllowedLanguages(dbLanguages []db.ShowAllowedLanguages) []structures.ShowAllowedLanguage {
	languages := make([]structures.ShowAllowedLanguage, len(dbLanguages))
	for i, l := range dbLanguages {
		languages[i] = structures.ShowAllowedLanguage{
			ID:           l.ID,
			LanguageCode: l.LanguageCode,
		}
	}
	return languages
}

func mapBlacklistedGenres(dbGenres []db.ShowBlacklistedGenres) []structures.BlacklistedShowGenre {
	genres := make([]structures.BlacklistedShowGenre, len(dbGenres))
	for i, g := range dbGenres {
		genres[i] = structures.BlacklistedShowGenre{
			ID:    g.ID,
			Genre: g.Genre,
		}
	}
	return genres
}

func mapBlacklistedNetworks(dbNetworks []db.ShowBlacklistedNetworks) []structures.BlacklistedNetwork {
	networks := make([]structures.BlacklistedNetwork, len(dbNetworks))
	for i, n := range dbNetworks {
		networks[i] = structures.BlacklistedNetwork{
			ID:      n.ID,
			Network: n.Network,
		}
	}
	return networks
}

func mapBlacklistedShowTitleKeywords(dbKeywords []db.ShowBlacklistedTitleKeywords) []structures.BlacklistedShowTitleKeyword {
	keywords := make([]structures.BlacklistedShowTitleKeyword, len(dbKeywords))
	for i, k := range dbKeywords {
		keywords[i] = structures.BlacklistedShowTitleKeyword{
			ID:      k.ID,
			Keyword: k.Keyword,
		}
	}
	return keywords
}

func mapBlacklistedTVDBIDs(dbTVDBIDs []db.ShowBlacklistedTVDBIDs) []structures.BlacklistedTVDBID {
	tmdbIDs := make([]structures.BlacklistedTVDBID, len(dbTVDBIDs))
	for i, t := range dbTVDBIDs {
		tmdbIDs[i] = structures.BlacklistedTVDBID{
			ID:     t.ID,
			TVDBID: t.TVDBID,
		}
	}
	return tmdbIDs
}
