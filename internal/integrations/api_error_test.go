package integrations

import (
	"context"
	"errors"
	"net/http"
	"net/url"
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

func TestTransportErrorsHideCredentialsAndPreserveCauses(t *testing.T) {
	for _, cause := range []error{context.Canceled, context.DeadlineExceeded, errors.New("connection failed")} {
		client := NewTMDB(TMDBConfig{APIKey: "synthetic-api-secret", SessionID: "synthetic-session-secret", AccountID: 1})
		transport := roundTripFunc(func(*http.Request) (*http.Response, error) { return nil, cause })
		client.httpClient.Transport = transport
		mdb := NewMDBList(MDBListConfig{APIKey: "synthetic-api-secret"})
		mdb.httpClient.Transport = transport
		checks := []func() error{
			func() error { return client.get(t.Context(), "/movie/popular", url.Values{}, &struct{}{}) },
			func() error { _, err := client.GetListItems(t.Context(), "", true, "movie", 1); return err },
			func() error { return mdb.Validate(t.Context()) },
		}
		for _, check := range checks {
			err := check()
			var urlErr *url.Error
			if !errors.Is(err, cause) || !errors.As(err, &urlErr) {
				t.Fatalf("transport cause lost: %v", err)
			}
			if strings.Contains(err.Error(), "synthetic-") {
				t.Fatal("transport error exposed credentials")
			}
		}
	}
}
