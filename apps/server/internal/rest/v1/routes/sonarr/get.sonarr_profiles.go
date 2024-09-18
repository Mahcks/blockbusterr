package sonarr

import (
	"errors"

	"github.com/mahcks/blockbusterr/internal/helpers/radarr"
	"github.com/mahcks/blockbusterr/internal/helpers/sonarr"
	"github.com/mahcks/blockbusterr/internal/rest/v1/respond"
	commonErrors "github.com/mahcks/blockbusterr/pkg/errors"
	"github.com/mahcks/blockbusterr/pkg/structures"
)

func (rg *RouteGroup) GetSonarrProfiles(ctx *respond.Ctx) error {
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

	profiles, err := rg.helpers.Sonarr.GetQualityProfiles(sonarrURL, apiKey)
	if err != nil {
		// Check if it's an unauthorized error and return 401 if true
		if errors.Is(err, radarr.ErrUnauthorizedRadarrRequest) {
			return commonErrors.ErrUnauthorized().SetDetail("Unauthorized access to Radarr")
		}

		return commonErrors.ErrInternalServerError().SetDetail("Failed to retrieve Sonarr profiles")
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
