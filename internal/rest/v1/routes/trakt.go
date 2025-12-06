package routes

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/mahcks/blockbusterr/internal/integrations"
)

// RegisterTraktRoutes registers Trakt-related API routes
func RegisterTraktRoutes(rg *RouteGroup, group fiber.Router) {
	trakt := group.Group("/trakt")

	// GET /v1/trakt/trending/movies?limit=10
	trakt.Get("/trending/movies", rg.GetTrendingMovies)

	// GET /v1/trakt/trending/shows?limit=10
	trakt.Get("/trending/shows", rg.GetTrendingShows)

	// GET /v1/trakt/popular/movies?limit=10
	trakt.Get("/popular/movies", rg.GetPopularMovies)

	// GET /v1/trakt/popular/shows?limit=10
	trakt.Get("/popular/shows", rg.GetPopularShows)

	// GET /v1/trakt/search?query=inception&type=movie&limit=10
	trakt.Get("/search", rg.SearchTrakt)

	// GET /v1/trakt/validate - Test connection
	trakt.Get("/validate", rg.ValidateTrakt)

	// Metadata endpoints for filters
	// GET /v1/trakt/languages/movies
	trakt.Get("/languages/:type", rg.GetLanguages)

	// GET /v1/trakt/genres/movies
	trakt.Get("/genres/:type", rg.GetGenres)

	// GET /v1/trakt/countries/movies
	trakt.Get("/countries/:type", rg.GetCountries)

	// GET /v1/trakt/networks
	trakt.Get("/networks", rg.GetNetworks)
}

// GetTrendingMovies returns trending movies from Trakt
func (rg *RouteGroup) GetTrendingMovies(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	cfg := rg.gctx.Config()
	traktClient := integrations.NewTrakt(integrations.TraktConfig{
		ClientID:     cfg.Trakt.ClientID,
		ClientSecret: cfg.Trakt.ClientSecret,
	})

	movies, err := traktClient.GetTrendingMovies(c.Context(), limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":  movies,
		"count": len(movies),
	})
}

// GetTrendingShows returns trending TV shows from Trakt
func (rg *RouteGroup) GetTrendingShows(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	cfg := rg.gctx.Config()
	traktClient := integrations.NewTrakt(integrations.TraktConfig{
		ClientID:     cfg.Trakt.ClientID,
		ClientSecret: cfg.Trakt.ClientSecret,
	})

	shows, err := traktClient.GetTrendingShows(c.Context(), limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":  shows,
		"count": len(shows),
	})
}

// GetPopularMovies returns popular movies from Trakt
func (rg *RouteGroup) GetPopularMovies(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	cfg := rg.gctx.Config()
	traktClient := integrations.NewTrakt(integrations.TraktConfig{
		ClientID:     cfg.Trakt.ClientID,
		ClientSecret: cfg.Trakt.ClientSecret,
	})

	movies, err := traktClient.GetPopularMovies(c.Context(), limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":  movies,
		"count": len(movies),
	})
}

// GetPopularShows returns popular TV shows from Trakt
func (rg *RouteGroup) GetPopularShows(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	cfg := rg.gctx.Config()
	traktClient := integrations.NewTrakt(integrations.TraktConfig{
		ClientID:     cfg.Trakt.ClientID,
		ClientSecret: cfg.Trakt.ClientSecret,
	})

	shows, err := traktClient.GetPopularShows(c.Context(), limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":  shows,
		"count": len(shows),
	})
}

// SearchTrakt performs a search on Trakt
func (rg *RouteGroup) SearchTrakt(c *fiber.Ctx) error {
	query := c.Query("query")
	if query == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "query parameter is required",
		})
	}

	searchType := c.Query("type", "movie")
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	cfg := rg.gctx.Config()
	traktClient := integrations.NewTrakt(integrations.TraktConfig{
		ClientID:     cfg.Trakt.ClientID,
		ClientSecret: cfg.Trakt.ClientSecret,
	})

	results, err := traktClient.Search(c.Context(), query, searchType, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":  results,
		"count": len(results),
	})
}

// ValidateTrakt validates the Trakt API connection
func (rg *RouteGroup) ValidateTrakt(c *fiber.Ctx) error {
	// Check if we're in test mode (testing form values)
	testMode := c.Query("test") == "true"

	var clientID, clientSecret string

	if testMode {
		// In test mode, require query parameters
		clientID = c.Query("client_id")
		clientSecret = c.Query("client_secret")
	} else {
		// Normal mode, use saved config
		cfg := rg.gctx.Config()
		clientID = cfg.Trakt.ClientID
		clientSecret = cfg.Trakt.ClientSecret
	}

	if clientID == "" || clientSecret == "" {
		return c.Status(400).JSON(fiber.Map{
			"error":     "Trakt client ID or client secret is not configured",
			"connected": false,
		})
	}

	traktClient := integrations.NewTrakt(integrations.TraktConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	})

	if err := traktClient.Validate(c.Context()); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":     err.Error(),
			"connected": false,
		})
	}

	return c.JSON(fiber.Map{
		"message":   "Trakt API connection successful",
		"connected": true,
	})
}

// GetLanguages returns available languages from Trakt
func (rg *RouteGroup) GetLanguages(c *fiber.Ctx) error {
	mediaType := c.Params("type") // movies or shows
	if mediaType != "movies" && mediaType != "shows" {
		return c.Status(400).JSON(fiber.Map{
			"error": "type must be 'movies' or 'shows'",
		})
	}

	cfg := rg.gctx.Config()
	traktClient := integrations.NewTrakt(integrations.TraktConfig{
		ClientID:     cfg.Trakt.ClientID,
		ClientSecret: cfg.Trakt.ClientSecret,
	})

	languages, err := traktClient.GetLanguages(c.Context(), mediaType)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"languages": languages,
		"count":     len(languages),
	})
}

// GetGenres returns available genres from Trakt
func (rg *RouteGroup) GetGenres(c *fiber.Ctx) error {
	mediaType := c.Params("type") // movies or shows
	if mediaType != "movies" && mediaType != "shows" {
		return c.Status(400).JSON(fiber.Map{
			"error": "type must be 'movies' or 'shows'",
		})
	}

	cfg := rg.gctx.Config()
	traktClient := integrations.NewTrakt(integrations.TraktConfig{
		ClientID:     cfg.Trakt.ClientID,
		ClientSecret: cfg.Trakt.ClientSecret,
	})

	genres, err := traktClient.GetGenres(c.Context(), mediaType)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"genres": genres,
		"count":  len(genres),
	})
}

// GetCountries returns available countries from Trakt
func (rg *RouteGroup) GetCountries(c *fiber.Ctx) error {
	mediaType := c.Params("type") // movies or shows
	if mediaType != "movies" && mediaType != "shows" {
		return c.Status(400).JSON(fiber.Map{
			"error": "type must be 'movies' or 'shows'",
		})
	}

	cfg := rg.gctx.Config()
	traktClient := integrations.NewTrakt(integrations.TraktConfig{
		ClientID:     cfg.Trakt.ClientID,
		ClientSecret: cfg.Trakt.ClientSecret,
	})

	countries, err := traktClient.GetCountries(c.Context(), mediaType)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"countries": countries,
		"count":     len(countries),
	})
}

// GetNetworks returns available TV networks from Trakt
func (rg *RouteGroup) GetNetworks(c *fiber.Ctx) error {
	cfg := rg.gctx.Config()
	traktClient := integrations.NewTrakt(integrations.TraktConfig{
		ClientID:     cfg.Trakt.ClientID,
		ClientSecret: cfg.Trakt.ClientSecret,
	})

	networks, err := traktClient.GetNetworks(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"networks": networks,
		"count":    len(networks),
	})
}
