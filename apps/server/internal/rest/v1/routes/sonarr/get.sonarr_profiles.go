package sonarr

import (
	"github.com/mahcks/blockbusterr/internal/helpers/sonarr"
	"github.com/mahcks/blockbusterr/internal/rest/v1/respond"
	"github.com/mahcks/blockbusterr/pkg/errors"
	"github.com/mahcks/blockbusterr/pkg/structures"
)

func (rg *RouteGroup) GetSonarrProfiles(ctx *respond.Ctx) error {
	profiles, err := rg.helpers.Sonarr.GetQualityProfiles()
	if err != nil {
		return errors.ErrInternalServerError().SetDetail("Failed to retrieve Sonarr profiles")
	}

	// Convert profiles to the desired structure
	var convertedProfiles []structures.SonarrQualityProfile
	for _, profile := range profiles {
		convertedProfile := convertQualityProfile(profile)
		convertedProfiles = append(convertedProfiles, convertedProfile)
	}

	// Respond with the converted profiles
	return ctx.JSON(convertedProfiles)
}

// convertQualityProfile converts the Sonarr API response to your SonarrQualityProfile structure
func convertQualityProfile(profile sonarr.QualityProfile) structures.SonarrQualityProfile {
	// Convert each quality item
	var items []structures.SonarrQualityItem
	for _, item := range profile.Items {
		items = append(items, convertQualityItem(item))
	}

	// Convert each format item
	var formatItems []structures.SonarrQualityItem
	for _, formatItem := range profile.FormatItems {
		formatItems = append(formatItems, convertQualityItem(formatItem))
	}

	return structures.SonarrQualityProfile{
		Name:              profile.Name,
		UpgradeAllowed:    profile.UpgradeAllowed,
		Cutoff:            profile.Cutoff,
		Items:             items,
		MinFormatScore:    profile.MinFormatScore,
		CutoffFormatScore: profile.CutoffFormatScore,
		FormatItems:       formatItems,
		ID:                profile.ID,
	}
}

// convertQualityItem converts a single Item to SonarrQualityItem
func convertQualityItem(item sonarr.Item) structures.SonarrQualityItem {
	var childItems []structures.SonarrQualityItem
	for _, child := range item.Items {
		childItems = append(childItems, convertQualityItem(child))
	}

	var quality *structures.SonarrQuality
	if item.Quality != nil {
		quality = &structures.SonarrQuality{
			ID:         item.Quality.ID,
			Name:       item.Quality.Name,
			Source:     item.Quality.Source,
			Resolution: item.Quality.Resolution,
		}
	}

	return structures.SonarrQualityItem{
		Quality: quality,
		Items:   childItems,
		Allowed: item.Allowed,
		Name:    item.Name,
		ID:      item.ID,
	}
}
