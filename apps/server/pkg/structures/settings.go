package structures

type Setting string

func (s Setting) String() string {
	return string(s)
}

const (
	// Trakt

	SettingTraktClientID     Setting = "TRAKT_CLIENT_ID"
	SettingTraktClientSecret Setting = "TRAKT_CLIENT_SECRET"

	// Ombi

	SettingOmbiURL    Setting = "OMBI_URL"
	SettingOmbiAPIKey Setting = "OMBI_API_KEY"

	// Radarr

	SettingRadarrURL    Setting = "RADARR_URL"
	SettingRadarrAPIKey Setting = "RADARR_API_KEY"

	// UI Flags

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
