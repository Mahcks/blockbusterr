package apiErrors

import (
	"fmt"
	"strings"

	"github.com/mahcks/blockbusterr/pkg/utils"
	"github.com/valyala/fasthttp"
)

type APIError interface {
	Error() string
	Message() string
	Code() int
	SetDetail(str string, a ...any) APIError
	SetFields(d Fields) APIError
	GetFields() Fields
	ExpectedHTTPStatus() int
	WithHTTPStatus(s int) APIError
}

type apiErrorFunc func() APIError

var (
	// Generic client errors
	ErrUnauthorized            apiErrorFunc = DefineError(10401, "Authorization Required", fasthttp.StatusUnauthorized)
	ErrTokenExpired            apiErrorFunc = DefineError(10402, "Token Expired", fasthttp.StatusUnauthorized)
	ErrInvalidToken            apiErrorFunc = DefineError(10403, "Invalid Token", fasthttp.StatusUnauthorized)
	ErrInsufficientPermissions apiErrorFunc = DefineError(10404, "Insufficient Permissions", fasthttp.StatusForbidden)
	ErrBadRequest              apiErrorFunc = DefineError(10405, "Bad Request", fasthttp.StatusBadRequest)
	ErrForbidden               apiErrorFunc = DefineError(10406, "Forbidden", fasthttp.StatusForbidden)
	ErrConflict                apiErrorFunc = DefineError(10407, "Conflict", fasthttp.StatusConflict)      // Used for conflicts like duplicate entries.
	ErrBadGateway              apiErrorFunc = DefineError(10408, "Bad Gateway", fasthttp.StatusBadGateway) // Used for upstream service errors.
	ErrTooManyRequests         apiErrorFunc = DefineError(10409, "Too Many Requests", fasthttp.StatusTooManyRequests)

	// Client type errors
	ErrValidationRejected         apiErrorFunc = DefineError(10410, "Validation Rejected", fasthttp.StatusBadRequest)
	ErrMissingEnvironmentVariable apiErrorFunc = DefineError(10411, "Missing Required Environment Variable", fasthttp.StatusBadRequest)

	// Server errors
	ErrInternalServerError apiErrorFunc = DefineError(10500, "Internal Server Error", fasthttp.StatusInternalServerError)
	ErrNotFound            apiErrorFunc = DefineError(10501, "Not Found", fasthttp.StatusNotFound)
	ErrInvalidSignature    apiErrorFunc = DefineError(10502, "Invalid Signature", fasthttp.StatusForbidden)
)

type apiError struct {
	message            string
	code               int
	fields             Fields
	expectedHTTPStatus int
}

type Fields map[string]any

func (e *apiError) Error() string {
	return fmt.Sprintf("[%d] %s", e.code, strings.ToLower(e.message))
}

func (e *apiError) Message() string {
	return e.message
}

func (e *apiError) Code() int {
	return e.code
}

func (e *apiError) SetDetail(str string, a ...any) APIError {
	e.message = e.message + ": " + utils.Ternary(len(a) > 0, fmt.Sprintf(str, a...), str)
	return e
}

func (e *apiError) SetFields(d Fields) APIError {
	e.fields = d
	return e
}

func (e *apiError) GetFields() Fields {
	return e.fields
}

func (e *apiError) ExpectedHTTPStatus() int {
	return e.expectedHTTPStatus
}

func (e *apiError) WithHTTPStatus(s int) APIError {
	e.expectedHTTPStatus = s
	return e
}

func DefineError(code int, message string, httpStatus int) func() APIError {
	return func() APIError {
		return &apiError{
			message:            message,
			code:               code,
			fields:             Fields{},
			expectedHTTPStatus: httpStatus,
		}
	}
}
