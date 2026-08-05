package integrations

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestAPIErrorDuplicateClassificationAndRedaction(t *testing.T) {
	duplicate := &APIError{Provider: "Radarr", StatusCode: http.StatusBadRequest, Code: "MovieExistsValidator"}
	if !IsDuplicateError(duplicate) || IsDuplicateError(errors.New("movie already exists")) || IsDuplicateError(&APIError{Provider: "Radarr", StatusCode: http.StatusInternalServerError, Code: "already"}) {
		t.Fatal("duplicate classification used message text or ignored provider status/code")
	}
	diagnostic := redactedDiagnostic([]byte(`{"message":"bad","apiKey":"secret","nested":{"password":"hidden"}}`))
	if strings.Contains(diagnostic, "secret") || strings.Contains(diagnostic, "hidden") {
		t.Fatalf("diagnostic was not redacted: %s", diagnostic)
	}
}
