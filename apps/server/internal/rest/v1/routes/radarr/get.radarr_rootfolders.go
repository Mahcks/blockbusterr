package radarr

import (
	"github.com/mahcks/blockbusterr/internal/helpers/radarr"
	"github.com/mahcks/blockbusterr/internal/rest/v1/respond"
	"github.com/mahcks/blockbusterr/pkg/errors"
	"github.com/mahcks/blockbusterr/pkg/structures"
)

func (rg *RouteGroup) GetRadarrRootFolders(ctx *respond.Ctx) error {
	// Fetch root folders from Radarr
	rootFolders, err := rg.helpers.Radarr.GetRootFolders()
	if err != nil {
		return errors.ErrInternalServerError().SetDetail("Failed to retrieve Radarr root folders")
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
