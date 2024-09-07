package structures

type Setting string

func (s Setting) String() string {
	return string(s)
}

const (
	// Trakt

	// SettingTraktClientID is the Trakt client ID
	SettingTraktClientID Setting = "TRAKT_CLIENT_ID"
	// SettingTraktClientSecret is the Trakt client secret
	SettingTraktClientSecret Setting = "TRAKT_CLIENT_SECRET"

	// Ombi

	// SettingOmbiURL is the Ombi URL
	SettingOmbiURL Setting = "OMBI_URL"
	// SettingOmbiAPIKey is the Ombi API key
	SettingOmbiAPIKey Setting = "OMBI_API_KEY"

	// Radarr

	// SettingRadarrURL is the Radarr URL
	SettingRadarrURL Setting = "RADARR_URL"
	// SettingRadarrAPIKey is the Radarr API key
	SettingRadarrAPIKey Setting = "RADARR_API_KEY"
	// SettingRadarrMinimumAvailability is the Radarr minimum availability
	SettingRadarrMinimumAvailability Setting = "RADARR_MINIMUM_AVAILABILITY"
	// SettingRadarrQuality is the quality of the movie
	SettingRadarrQuality Setting = "RADARR_QUALITY"
	// SettingRadarrRootFolderPath is the Radarr root folder path
	SettingRadarrRootFolderPath Setting = "RADARR_ROOT_FOLDER_PATH"

	// UI Flags

	// SettingSetupComplete is a flag to indicate if the setup is complete
	SettingSetupComplete Setting = "SETUP_COMPLETE"
)

func IsValidSettingKey(key Setting) bool {
	switch key {
	case SettingTraktClientID,
		SettingTraktClientSecret,
		SettingSetupComplete:
		return true
	default:
		return false
	}
}
