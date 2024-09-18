package sonarr

import (
	"errors"

	"github.com/mahcks/blockbusterr/internal/helpers/radarr"
	"github.com/mahcks/blockbusterr/internal/rest/v1/respond"
	commonErrors "github.com/mahcks/blockbusterr/pkg/errors"
	"github.com/mahcks/blockbusterr/pkg/structures"
)

func (rg *RouteGroup) GetSonarrRootFolders(ctx *respond.Ctx) error {
	// Extract Sonarr URL and API key from the query params and headers
	var sonarrURL *string
	if url := ctx.Query("url"); url != "" {
		sonarrURL = &url
	}

	var apiKey *string
	headers := ctx.GetReqHeaders()
	if apiKeyArray, exists := headers["X-Api-Key"]; exists && len(apiKeyArray) > 0 {
		apiKey = &apiKeyArray[0]
	}

	sonarrFolders, err := rg.helpers.Sonarr.GetRootFolders(sonarrURL, apiKey)
	if err != nil {
		// Check if it's an unauthorized error and return 401 if true
		if errors.Is(err, radarr.ErrUnauthorizedRadarrRequest) {
			return commonErrors.ErrUnauthorized().SetDetail("Unauthorized access to Radarr")
		}

		return commonErrors.ErrInternalServerError().SetDetail("Failed to retrieve Sonarr root folders")
	}

	// Convert Sonarr root folders to the desired structure
	var convertedFolders []structures.SonarrRootFolder
	for _, folder := range sonarrFolders {
		convertedFolders = append(convertedFolders, structures.SonarrRootFolder{
			ID:         folder.ID,
			Path:       folder.Path,
			Accessible: folder.Accessible,
			FreeSpace:  folder.FreeSpace,
		})
	}

	return ctx.JSON(convertedFolders)
}
