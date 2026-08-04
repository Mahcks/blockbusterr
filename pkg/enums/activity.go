package enums

import "strings"

type MediaType string

const (
	MediaTypeMovie MediaType = "movie"
	MediaTypeShow  MediaType = "show"
)

func (m MediaType) IsValid() bool {
	switch m {
	case MediaTypeMovie, MediaTypeShow:
		return true
	default:
		return false
	}
}

func ParseMediaType(value string) (MediaType, bool) {
	normalized := MediaType(strings.ToLower(strings.TrimSpace(value)))
	if normalized.IsValid() {
		return normalized, true
	}
	return "", false
}

type ActivityStatus string

const (
	ActivityStatusAdded     ActivityStatus = "added"
	ActivityStatusRequested ActivityStatus = "requested"
	ActivityStatusRejected  ActivityStatus = "rejected"
	ActivityStatusSkipped   ActivityStatus = "skipped"
	ActivityStatusFailed    ActivityStatus = "failed"
	ActivityStatusBlocked   ActivityStatus = "blocked"
)

func (s ActivityStatus) IsValid() bool {
	switch s {
	case ActivityStatusAdded,
		ActivityStatusRequested,
		ActivityStatusRejected,
		ActivityStatusSkipped,
		ActivityStatusFailed,
		ActivityStatusBlocked:
		return true
	default:
		return false
	}
}

func ParseActivityStatus(value string) (ActivityStatus, bool) {
	normalized := ActivityStatus(strings.ToLower(strings.TrimSpace(value)))
	if normalized.IsValid() {
		return normalized, true
	}
	return "", false
}

func IsSuccessLikeActivityStatus(s ActivityStatus) bool {
	return s == ActivityStatusAdded || s == ActivityStatusRequested
}

type RepeatPolicy string

const (
	RepeatPolicyInherit   RepeatPolicy = ""
	RepeatPolicyImmediate RepeatPolicy = "immediate"
	RepeatPolicy30Days    RepeatPolicy = "30_days"
	RepeatPolicy90Days    RepeatPolicy = "90_days"
	RepeatPolicy180Days   RepeatPolicy = "180_days"
	RepeatPolicyNever     RepeatPolicy = "never"
)

func (p RepeatPolicy) IsValid(allowInherit bool) bool {
	if allowInherit && p == RepeatPolicyInherit {
		return true
	}
	switch p {
	case RepeatPolicyImmediate, RepeatPolicy30Days, RepeatPolicy90Days, RepeatPolicy180Days, RepeatPolicyNever:
		return true
	default:
		return false
	}
}

type JobRunStatus string

const (
	JobRunStatusRunning   JobRunStatus = "running"
	JobRunStatusCompleted JobRunStatus = "completed"
	JobRunStatusFailed    JobRunStatus = "failed"
)

func (s JobRunStatus) IsValid() bool {
	switch s {
	case JobRunStatusRunning, JobRunStatusCompleted, JobRunStatusFailed:
		return true
	default:
		return false
	}
}

func ParseJobRunStatus(value string) (JobRunStatus, bool) {
	normalized := JobRunStatus(strings.ToLower(strings.TrimSpace(value)))
	if normalized.IsValid() {
		return normalized, true
	}
	return "", false
}
