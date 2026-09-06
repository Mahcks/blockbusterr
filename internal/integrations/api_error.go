package integrations

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

const upstreamErrorBodyLimit = 4 << 10

type APIError struct {
	Provider   string
	StatusCode int
	Code       string
}

func (err *APIError) Error() string {
	return fmt.Sprintf("%s request failed with status %d", err.Provider, err.StatusCode)
}

func IsDuplicateError(err error) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	return apiErr.StatusCode == http.StatusConflict || apiErr.Code == "MovieExistsValidator" || apiErr.Code == "SeriesExistsValidator"
}

func newAPIError(provider string, response *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(response.Body, upstreamErrorBodyLimit))
	code := providerErrorCode(body)
	slog.Warn("Upstream API request failed", "provider", provider, "status", response.StatusCode, "code", code, "body", redactedDiagnostic(body))
	return &APIError{Provider: provider, StatusCode: response.StatusCode, Code: code}
}

func redactedDiagnostic(body []byte) string {
	var value any
	if json.Unmarshal(body, &value) != nil {
		return "<non-JSON response omitted>"
	}
	redactValue(value)
	redacted, _ := json.Marshal(value)
	return string(redacted)
}

func redactValue(value any) {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			lower := strings.ToLower(key)
			if strings.Contains(lower, "token") || strings.Contains(lower, "secret") || strings.Contains(lower, "password") || strings.Contains(lower, "api_key") || strings.Contains(lower, "apikey") {
				typed[key] = "[REDACTED]"
				continue
			}
			redactValue(child)
		}
	case []any:
		for _, child := range typed {
			redactValue(child)
		}
	}
}

func providerErrorCode(body []byte) string {
	var list []struct {
		Code string `json:"errorCode"`
	}
	if json.Unmarshal(body, &list) == nil && len(list) > 0 {
		return list[0].Code
	}
	var item struct {
		Code string `json:"errorCode"`
	}
	_ = json.Unmarshal(body, &item)
	return item.Code
}

// requestError keeps credential-bearing request URLs out of logs and history.
// Unwrap retains cancellation, timeout, and transport error inspection.
type requestError struct{ cause error }

func (err *requestError) Error() string {
	if errors.Is(err.cause, context.Canceled) {
		return "HTTP request canceled"
	}
	if errors.Is(err.cause, context.DeadlineExceeded) {
		return "HTTP request deadline exceeded"
	}
	return "HTTP transport failed"
}
func (err *requestError) Unwrap() error { return err.cause }

func doRequest(client *http.Client, req *http.Request) (*http.Response, error) {
	response, err := client.Do(req)
	if err != nil {
		return response, &requestError{cause: err}
	}
	return response, nil
}
