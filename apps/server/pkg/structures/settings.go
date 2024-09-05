package structures

type Setting string

const (
	// Trakt

	SettingTraktClientID     Setting = "TRAKT_CLIENT_ID"
	SettingTraktClientSecret Setting = "TRAKT_CLIENT_SECRET"

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
