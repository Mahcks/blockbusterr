package config

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

// ParseSyncInterval accepts only schedules that can produce a positive wait.
func ParseSyncInterval(interval string) (time.Duration, cron.Schedule, error) {
	if duration, err := time.ParseDuration(interval); err == nil {
		if duration > 0 {
			return duration, nil, nil
		}
		return 0, nil, fmt.Errorf("interval must be positive")
	}
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(interval)
	if err != nil || schedule.Next(time.Now()).IsZero() {
		return 0, nil, fmt.Errorf("interval must be a positive duration or five-field cron expression with a future occurrence")
	}
	return 0, schedule, nil
}
