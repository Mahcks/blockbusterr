package validation

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Validator interface for all validators
type Validator interface {
	Validate(value interface{}) error
}

// URLValidator validates URL strings
type URLValidator struct {
	Required bool
	Schemes  []string // allowed schemes (http, https)
}

func (v *URLValidator) Validate(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return &ValidationError{Field: "url", Message: "must be a string"}
	}

	str = strings.TrimSpace(str)
	if str == "" {
		if v.Required {
			return &ValidationError{Field: "url", Message: "is required"}
		}
		return nil
	}

	parsedURL, err := url.Parse(str)
	if err != nil {
		return &ValidationError{Field: "url", Message: "is not a valid URL"}
	}

	if parsedURL.Scheme == "" {
		return &ValidationError{Field: "url", Message: "must include a scheme (http or https)"}
	}

	if len(v.Schemes) > 0 {
		validScheme := false
		for _, scheme := range v.Schemes {
			if parsedURL.Scheme == scheme {
				validScheme = true
				break
			}
		}
		if !validScheme {
			return &ValidationError{
				Field:   "url",
				Message: fmt.Sprintf("scheme must be one of: %s", strings.Join(v.Schemes, ", ")),
			}
		}
	}

	if parsedURL.Host == "" {
		return &ValidationError{Field: "url", Message: "must include a host"}
	}

	return nil
}

// APIKeyValidator validates API key strings
type APIKeyValidator struct {
	Required  bool
	MinLength int
	MaxLength int
}

func (v *APIKeyValidator) Validate(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return &ValidationError{Field: "api_key", Message: "must be a string"}
	}

	str = strings.TrimSpace(str)
	if str == "" {
		if v.Required {
			return &ValidationError{Field: "api_key", Message: "is required"}
		}
		return nil
	}

	if v.MinLength > 0 && len(str) < v.MinLength {
		return &ValidationError{
			Field:   "api_key",
			Message: fmt.Sprintf("must be at least %d characters", v.MinLength),
		}
	}

	if v.MaxLength > 0 && len(str) > v.MaxLength {
		return &ValidationError{
			Field:   "api_key",
			Message: fmt.Sprintf("must be at most %d characters", v.MaxLength),
		}
	}

	// Check if it's alphanumeric
	matched, err := regexp.MatchString("^[a-zA-Z0-9]+$", str)
	if err != nil || !matched {
		return &ValidationError{Field: "api_key", Message: "must contain only alphanumeric characters"}
	}

	return nil
}

// IntervalValidator validates interval strings (e.g., "1h", "30m")
type IntervalValidator struct {
	Required    bool
	MinDuration int // in minutes
	MaxDuration int // in minutes
}

func (v *IntervalValidator) Validate(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return &ValidationError{Field: "interval", Message: "must be a string"}
	}

	str = strings.TrimSpace(str)
	if str == "" {
		if v.Required {
			return &ValidationError{Field: "interval", Message: "is required"}
		}
		return nil
	}

	// Parse interval format: number + unit (m=minutes, h=hours)
	matched, err := regexp.MatchString("^[0-9]+[mh]$", str)
	if err != nil || !matched {
		return &ValidationError{Field: "interval", Message: "must be in format like '30m' or '1h'"}
	}

	// Extract number and unit
	unit := str[len(str)-1:]
	numStr := str[:len(str)-1]
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return &ValidationError{Field: "interval", Message: "invalid number in interval"}
	}

	// Convert to minutes
	minutes := num
	if unit == "h" {
		minutes = num * 60
	}

	if v.MinDuration > 0 && minutes < v.MinDuration {
		return &ValidationError{
			Field:   "interval",
			Message: fmt.Sprintf("must be at least %d minutes", v.MinDuration),
		}
	}

	if v.MaxDuration > 0 && minutes > v.MaxDuration {
		return &ValidationError{
			Field:   "interval",
			Message: fmt.Sprintf("must be at most %d minutes", v.MaxDuration),
		}
	}

	return nil
}

// LimitValidator validates content limit integers
type LimitValidator struct {
	Required bool
	Min      int
	Max      int
}

func (v *LimitValidator) Validate(value interface{}) error {
	var num int

	switch val := value.(type) {
	case int:
		num = val
	case string:
		if val == "" {
			if v.Required {
				return &ValidationError{Field: "limit", Message: "is required"}
			}
			return nil
		}
		var err error
		num, err = strconv.Atoi(val)
		if err != nil {
			return &ValidationError{Field: "limit", Message: "must be a valid number"}
		}
	default:
		return &ValidationError{Field: "limit", Message: "must be a number"}
	}

	if v.Min > 0 && num < v.Min {
		return &ValidationError{
			Field:   "limit",
			Message: fmt.Sprintf("must be at least %d", v.Min),
		}
	}

	if v.Max > 0 && num > v.Max {
		return &ValidationError{
			Field:   "limit",
			Message: fmt.Sprintf("must be at most %d", v.Max),
		}
	}

	return nil
}

// PeriodValidator validates period strings (weekly, monthly, yearly, all)
type PeriodValidator struct {
	Required bool
}

func (v *PeriodValidator) Validate(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return &ValidationError{Field: "period", Message: "must be a string"}
	}

	str = strings.TrimSpace(strings.ToLower(str))
	if str == "" {
		if v.Required {
			return &ValidationError{Field: "period", Message: "is required"}
		}
		return nil
	}

	validPeriods := []string{"weekly", "monthly", "yearly", "all"}
	for _, period := range validPeriods {
		if str == period {
			return nil
		}
	}

	return &ValidationError{
		Field:   "period",
		Message: fmt.Sprintf("must be one of: %s", strings.Join(validPeriods, ", ")),
	}
}

func (v *PeriodValidator) GetValidPeriods() []string {
	return []string{"weekly", "monthly", "yearly", "all"}
}
