package media

import (
	"github.com/mahcks/blockbusterr/internal/rest/v1/respond"
	"github.com/mahcks/blockbusterr/pkg/structures"
)

func (rg *RouteGroup) GetMediaRecentlyAdded(ctx *respond.Ctx) error {
	var media []structures.RecentlyAddedMedia

	// Default values for pagination
	page := ctx.QueryInt("page", 1)          // Defaults to page 1 if not provided
	pageSize := ctx.QueryInt("pageSize", 10) // Defaults to 10 items per page if not provided

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// Calculate limit and offset based on page and pageSize
	limit := pageSize
	offset := (page - 1) * pageSize

	// Get the recently added media with pagination
	recentlyAdded, err := rg.gctx.Crate().SQL.Queries().GetRecentlyAddedMedia(ctx.Context(), limit, offset)
	if err != nil {
		return err
	}

	for _, recentlyAddedMedia := range recentlyAdded {
		media = append(media, structures.RecentlyAddedMedia{
			ID:        recentlyAddedMedia.ID,
			Title:     recentlyAddedMedia.Title,
			MediaType: recentlyAddedMedia.MediaType,
			Summary:   recentlyAddedMedia.Summary,
			IMDBID:    recentlyAddedMedia.IMDBID,
			Year:      recentlyAddedMedia.Year,
			Poster:    recentlyAddedMedia.Poster,
			AddedAt:   recentlyAddedMedia.AddedAt,
		})
	}

	// Respond with paginated media
	return ctx.JSON(media)
}
