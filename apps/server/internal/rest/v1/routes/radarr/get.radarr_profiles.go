package radarr

import (
	"github.com/mahcks/blockbusterr/internal/helpers/radarr"
	"github.com/mahcks/blockbusterr/internal/rest/v1/respond"
	"github.com/mahcks/blockbusterr/pkg/errors"
	"github.com/mahcks/blockbusterr/pkg/structures"
)

func (rg *RouteGroup) GetRadarrProfiles(ctx *respond.Ctx) error {
	// Fetch quality profiles from Radarr
	profiles, err := rg.helpers.Radarr.GetQualityProfiles()
	if err != nil {
		return errors.ErrInternalServerError().SetDetail("Failed to retrieve Radarr profiles")
	}

	// Convert Radarr profiles to structures.QualityProfile
	var convertedProfiles []structures.RadarrQualityProfile
	for _, profile := range profiles {
		convertedProfile := structures.RadarrQualityProfile{
			ID:                profile.ID,
			Name:              profile.Name,
			UpgradeAllowed:    profile.UpgradeAllowed,
			Cutoff:            profile.Cutoff,
			MinFormatScore:    profile.MinFormatScore,
			CutoffFormatScore: profile.CutoffFormatScore,
			Items:             convertQualityProfileItems(profile.Items), // Conversion for nested structs
			Language:          convertQualityProfileLanguage(profile.Language),
		}

		// Add the converted profile to the result slice
		convertedProfiles = append(convertedProfiles, convertedProfile)
	}

	// Return the converted profiles in the response
	return ctx.JSON(convertedProfiles)
}

func convertQualityProfileItems(items radarr.QualityProfileItem) structures.RadarrQualityProfileItem {
	convertedItems := make(structures.RadarrQualityProfileItem, len(items))
	for i, item := range items {
		convertedItems[i] = structures.RadarrQualityProfileItemData{
			Quality: structures.RadarrQuality{
				ID:         item.Quality.ID,
				Name:       item.Quality.Name,
				Source:     item.Quality.Source,
				Resolution: item.Quality.Resolution,
				Modifier:   item.Quality.Modifier,
			},
			Allowed: item.Allowed,
		}
	}
	return convertedItems
}

func convertQualityProfileLanguage(lang radarr.QualityProfileLanguage) structures.RadarrQualityProfileLanguage {
	return structures.RadarrQualityProfileLanguage{
		ID:   lang.ID,
		Name: lang.Name,
	}
}
