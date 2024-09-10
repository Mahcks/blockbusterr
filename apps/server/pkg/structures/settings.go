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

	// UI Flags

	// SettingSetupComplete is a flag to indicate if the setup is complete
	SettingSetupComplete Setting = "SETUP_COMPLETE"
	SettingOmbiEnabled   Setting = "OMBI_ENABLED"
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
