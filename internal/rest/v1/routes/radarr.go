package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mahcks/blockbusterr/internal/integrations"
)

// RegisterRadarrRoutes registers Radarr-related API routes
func RegisterRadarrRoutes(rg *RouteGroup, group fiber.Router) {
	radarr := group.Group("/radarr")

	// GET /v1/radarr/validate - Test connection
	radarr.Get("/validate", rg.ValidateRadarr)
	radarr.Post("/validate", rg.ValidateRadarr)

	// GET /v1/radarr/movies - List all movies
	radarr.Get("/movies", rg.GetRadarrMovies)

	// GET /v1/radarr/quality-profiles - Get quality profiles
	radarr.Get("/quality-profiles", rg.GetRadarrQualityProfiles)

	// GET /v1/radarr/root-folders - Get root folders
	radarr.Get("/root-folders", rg.GetRadarrRootFolders)

	// GET /v1/radarr/lookup?term=inception - Lookup movie
	radarr.Get("/lookup", rg.LookupRadarrMovie)

	// POST /v1/radarr/movies - Add a movie
	radarr.Post("/movies", rg.AddRadarrMovie)
}

// ValidateRadarr validates the Radarr API connection
func (rg *RouteGroup) ValidateRadarr(c *fiber.Ctx) error {
	var url, apiKey string
	if c.Method() == fiber.MethodPost {
		var request struct {
			URL    string `json:"url"`
			APIKey string `json:"api_key"`
		}
		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body", "connected": false})
		}
		url, apiKey = request.URL, request.APIKey
	} else {
		// Normal mode, use saved config
		cfg := rg.gctx.Config()
		url = cfg.Radarr.URL
		apiKey = cfg.Radarr.APIKey
	}
	c.Set(fiber.HeaderCacheControl, "no-store")

	if url == "" || apiKey == "" {
		return c.Status(400).JSON(fiber.Map{
			"error":     "Radarr URL or API key is not configured",
			"connected": false,
		})
	}

	radarrClient := integrations.NewRadarr(integrations.RadarrConfig{
		BaseURL: url,
		APIKey:  apiKey,
	})

	status, err := radarrClient.GetSystemStatus(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":     err.Error(),
			"connected": false,
		})
	}

	return c.JSON(fiber.Map{
		"message":   "Radarr API connection successful",
		"connected": true,
		"version":   status.Version,
	})
}

// GetRadarrMovies returns all movies in Radarr
func (rg *RouteGroup) GetRadarrMovies(c *fiber.Ctx) error {
	cfg := rg.gctx.Config()

	radarrClient := integrations.NewRadarr(integrations.RadarrConfig{
		BaseURL: cfg.Radarr.URL,
		APIKey:  cfg.Radarr.APIKey,
	})

	movies, err := radarrClient.GetMovies(c.Context())
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

// GetRadarrQualityProfiles returns quality profiles
func (rg *RouteGroup) GetRadarrQualityProfiles(c *fiber.Ctx) error {
	cfg := rg.gctx.Config()

	radarrClient := integrations.NewRadarr(integrations.RadarrConfig{
		BaseURL: cfg.Radarr.URL,
		APIKey:  cfg.Radarr.APIKey,
	})

	profiles, err := radarrClient.GetQualityProfiles(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":  profiles,
		"count": len(profiles),
	})
}

// GetRadarrRootFolders returns root folders
func (rg *RouteGroup) GetRadarrRootFolders(c *fiber.Ctx) error {
	cfg := rg.gctx.Config()

	radarrClient := integrations.NewRadarr(integrations.RadarrConfig{
		BaseURL: cfg.Radarr.URL,
		APIKey:  cfg.Radarr.APIKey,
	})

	folders, err := radarrClient.GetRootFolders(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":  folders,
		"count": len(folders),
	})
}

// LookupRadarrMovie looks up a movie by term
func (rg *RouteGroup) LookupRadarrMovie(c *fiber.Ctx) error {
	term := c.Query("term")
	if term == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "term parameter is required",
		})
	}

	cfg := rg.gctx.Config()

	radarrClient := integrations.NewRadarr(integrations.RadarrConfig{
		BaseURL: cfg.Radarr.URL,
		APIKey:  cfg.Radarr.APIKey,
	})

	movies, err := radarrClient.LookupMovie(c.Context(), term)
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

// AddRadarrMovie adds a movie to Radarr
func (rg *RouteGroup) AddRadarrMovie(c *fiber.Ctx) error {
	var movie integrations.RadarrMovie
	if err := c.BodyParser(&movie); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	cfg := rg.gctx.Config()

	radarrClient := integrations.NewRadarr(integrations.RadarrConfig{
		BaseURL: cfg.Radarr.URL,
		APIKey:  cfg.Radarr.APIKey,
	})

	addedMovie, err := radarrClient.AddMovie(c.Context(), movie)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"data":    addedMovie,
		"message": "Movie added successfully",
	})
}
