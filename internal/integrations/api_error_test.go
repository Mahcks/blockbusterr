package integrations

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
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
	var output bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&output, nil)))
	defer slog.SetDefault(previous)
	err := newAPIError("Radarr", jsonResponse(http.StatusBadRequest, `{"message":"synthetic-secret","headers":{"Authorization":"synthetic-secret"},"errorCode":"synthetic-secret"}`))
	if strings.Contains(output.String(), "synthetic-secret") || strings.Contains(err.Error(), "synthetic-secret") {
		t.Fatal("upstream diagnostic exposed credentials")
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
