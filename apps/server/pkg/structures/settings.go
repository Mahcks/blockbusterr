package structures

type Setting string

func (s Setting) String() string {
	return string(s)
}

const (
	// Ombi

	// SettingSetupComplete is a flag to indicate if the setup is complete
	SettingSetupComplete Setting = "SETUP_COMPLETE"
	SettingMode          Setting = "MODE"
)

func IsValidSettingKey(key Setting) bool {
	switch key {
	case SettingSetupComplete, SettingMode:
		return true
	default:
		return false
	}
}
