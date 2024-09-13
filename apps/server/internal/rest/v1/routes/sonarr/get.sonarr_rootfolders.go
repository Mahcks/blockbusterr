package sonarr

import (
	"github.com/mahcks/blockbusterr/internal/rest/v1/respond"
	"github.com/mahcks/blockbusterr/pkg/errors"
	"github.com/mahcks/blockbusterr/pkg/structures"
)

func (rg *RouteGroup) GetSonarrRootFolders(ctx *respond.Ctx) error {
	sonarrFolders, err := rg.helpers.Sonarr.GetRootFolders()
	if err != nil {
		return errors.ErrInternalServerError().SetDetail("Failed to retrieve Sonarr root folders")
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
