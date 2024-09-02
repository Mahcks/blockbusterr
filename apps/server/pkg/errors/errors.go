package errors

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
	ErrInsufficientPermissions apiErrorFunc = DefineError(10403, "Insufficient Permissions", fasthttp.StatusForbidden)
	ErrBadRequest              apiErrorFunc = DefineError(10404, "Bad Request", fasthttp.StatusBadRequest)

	// Client type errors
	ErrValidationRejected apiErrorFunc = DefineError(10410, "Validation Rejected", fasthttp.StatusBadRequest)

	// Other client errors

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

type Fields map[string]interface{}

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
