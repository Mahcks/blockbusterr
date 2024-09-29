package structures

type Log struct {
	ID        int      `json:"id"`
	Label     string   `json:"label"`
	Level     LogLevel `json:"level"`
	Message   string   `json:"message"`
	Timestamp string   `json:"timestamp"`
}

type LogLevel string

func (ll LogLevel) String() string {
	return string(ll)
}

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelError LogLevel = "error"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
)
