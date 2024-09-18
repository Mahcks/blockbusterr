package radarr

import (
	"errors"

	"github.com/mahcks/blockbusterr/internal/helpers/radarr"
	"github.com/mahcks/blockbusterr/internal/rest/v1/respond"
	commonErrors "github.com/mahcks/blockbusterr/pkg/errors"
	"github.com/mahcks/blockbusterr/pkg/structures"
)

func (rg *RouteGroup) GetRadarrRootFolders(ctx *respond.Ctx) error {
	// Extract Radarr URL and API key from the query params and headers
	var radarrURL *string
	if url := ctx.Query("url"); url != "" {
		radarrURL = &url
	}

	var apiKey *string
	headers := ctx.GetReqHeaders()
	if apiKeyArray, exists := headers["X-Api-Key"]; exists && len(apiKeyArray) > 0 {
		apiKey = &apiKeyArray[0]
	}

	// Fetch root folders from Radarr
	rootFolders, err := rg.helpers.Radarr.GetRootFolders(radarrURL, apiKey)
	if err != nil {
		// Check if it's an unauthorized error and return 401 if true
		if errors.Is(err, radarr.ErrUnauthorizedRadarrRequest) {
			return commonErrors.ErrUnauthorized().SetDetail("Unauthorized access to Radarr")
		}

		return commonErrors.ErrInternalServerError().SetDetail("Failed to retrieve Radarr root folders")
	}

	// Convert Radarr root folders to structures.RadarrRootFolder
	var convertedFolders []structures.RadarrRootFolder
	for _, root := range rootFolders {
		convertedFolder := structures.RadarrRootFolder{
			ID:              root.ID,
			Path:            root.Path,
			Accessible:      root.Accessible,
			FreeSpace:       root.FreeSpace,
			UnmappedFolders: convertUnmappedFolders(root.UnmappedFolders), // Convert nested unmapped folders
		}

		// Add the converted folder to the result slice
		convertedFolders = append(convertedFolders, convertedFolder)
	}

	// Return the converted folders in the response
	return ctx.JSON(convertedFolders)
}

func convertUnmappedFolders(unmappedFolders []radarr.RootFolderUnmappedFolder) []structures.RadarrRootFolderUnmappedFolder {
	convertedFolders := make([]structures.RadarrRootFolderUnmappedFolder, len(unmappedFolders))
	for i, folder := range unmappedFolders {
		convertedFolders[i] = structures.RadarrRootFolderUnmappedFolder{
			Name:         folder.Name,
			Path:         folder.Path,
			RelativePath: folder.RelativePath,
		}
	}
	return convertedFolders
}
