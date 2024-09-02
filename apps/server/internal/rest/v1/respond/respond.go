package respond

import (
	fiber "github.com/gofiber/fiber/v2"
)

type Ctx struct {
	*fiber.Ctx
}

type APIErrorResponseBodyError struct {
	StatusCode int      `json:"status_code"`
	Timestamp  int      `json:"timestamp"`
	Error      APIError `json:"error"`
	TraceID    string   `json:"trace_id,omitempty"`
}

type APIError struct {
	StatusCode int                    `json:"status_code"`
	Error      string                 `json:"error"`
	ErrorCode  int                    `json:"error_code"`
	Details    map[string]interface{} `json:"details,omitempty"`
}
