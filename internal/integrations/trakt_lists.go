package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const TraktAuthBaseURL = "https://auth.trakt.tv"

type TraktListItems struct {
	Movies []Movie
	Shows  []Show
}

type TraktDeviceCode struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURL string `json:"verification_url"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

type TraktToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	CreatedAt    int64  `json:"created_at"`
}

func (token TraktToken) ExpiresAt() int64 { return token.CreatedAt + token.ExpiresIn }

func (t *Trakt) GetListItems(ctx context.Context, owner, listID string, watchlist bool, mediaType string, limit int) (TraktListItems, error) {
	var endpoint string
	if watchlist && t.accessToken != "" && (owner == "" || owner == "me") {
		endpoint = "/sync/watchlist/movie,show/rank/asc?extended=full"
	} else if watchlist {
		endpoint = fmt.Sprintf("/users/%s/watchlist/movie,show/rank/asc?extended=full", url.PathEscape(owner))
	} else if owner != "" {
		endpoint = fmt.Sprintf("/users/%s/lists/%s/items/movie,show?extended=full", url.PathEscape(owner), url.PathEscape(listID))
	} else {
		endpoint = fmt.Sprintf("/lists/%s/items/movie,show?extended=full", url.PathEscape(listID))
	}
	result := TraktListItems{}
	for page := 1; listMediaCount(mediaType, len(result.Movies), len(result.Shows)) < limit; page++ {
		separator := "?"
		if strings.Contains(endpoint, "?") {
			separator = "&"
		}
		response, err := t.doRequest(ctx, http.MethodGet, fmt.Sprintf("%s%spage=%d&limit=%d", endpoint, separator, page, TraktDefaultPageSize), nil)
		if err != nil {
			return TraktListItems{}, err
		}
		if response.StatusCode != http.StatusOK {
			_ = response.Body.Close()
			return TraktListItems{}, fmt.Errorf("Trakt API returned status %d", response.StatusCode)
		}
		var items []struct {
			Type  string `json:"type"`
			Movie *Movie `json:"movie"`
			Show  *Show  `json:"show"`
		}
		if err := json.NewDecoder(response.Body).Decode(&items); err != nil {
			_ = response.Body.Close()
			return TraktListItems{}, fmt.Errorf("failed to decode Trakt list: %w", err)
		}
		_ = response.Body.Close()
		for _, item := range items {
			if item.Type == "movie" && mediaType != "show" && item.Movie != nil && len(result.Movies) < limit {
				result.Movies = append(result.Movies, *item.Movie)
			} else if item.Type == "show" && mediaType != "movie" && item.Show != nil && len(result.Shows) < limit {
				result.Shows = append(result.Shows, *item.Show)
			}
		}
		if len(items) < TraktDefaultPageSize {
			break
		}
	}
	return result, nil
}

func (t *Trakt) StartDeviceAuth(ctx context.Context) (TraktDeviceCode, error) {
	var result TraktDeviceCode
	err := t.authJSON(ctx, "/oauth/device/code", map[string]string{"client_id": t.clientID}, &result)
	return result, err
}

func (t *Trakt) PollDeviceAuth(ctx context.Context, deviceCode string) (TraktToken, int, error) {
	var token TraktToken
	status, err := t.authJSONStatus(ctx, "/oauth/device/token", map[string]string{"code": deviceCode, "client_id": t.clientID, "client_secret": t.clientSecret}, &token)
	if err != nil || status != http.StatusOK {
		return token, status, err
	}
	if token.CreatedAt == 0 {
		token.CreatedAt = time.Now().Unix()
	}
	traktTokenMu.Lock()
	defer traktTokenMu.Unlock()
	return token, status, t.storeToken(token)
}

func (t *Trakt) accessTokenForRequest(ctx context.Context) (string, error) {
	traktTokenMu.Lock()
	defer traktTokenMu.Unlock()
	if t.loadToken != nil {
		token := t.loadToken()
		if token.AccessToken != "" {
			t.accessToken, t.refreshToken, t.tokenExpires = token.AccessToken, token.RefreshToken, token.ExpiresAt()
		}
	}
	if t.refreshToken == "" || t.tokenExpires <= 0 {
		return t.accessToken, nil
	}
	if time.Now().Unix() < t.tokenExpires-60 {
		return t.accessToken, nil
	}
	var token TraktToken
	status, err := t.authJSONStatus(ctx, "/oauth/token", map[string]string{"refresh_token": t.refreshToken, "client_id": t.clientID, "client_secret": t.clientSecret, "grant_type": "refresh_token", "redirect_uri": "urn:ietf:wg:oauth:2.0:oob"}, &token)
	if err != nil {
		return "", err
	}
	if status != http.StatusOK {
		return "", fmt.Errorf("Trakt token refresh returned status %d", status)
	}
	if token.CreatedAt == 0 {
		token.CreatedAt = time.Now().Unix()
	}
	if err := t.storeToken(token); err != nil {
		return "", err
	}
	return t.accessToken, nil
}

func (t *Trakt) storeToken(token TraktToken) error {
	t.accessToken, t.refreshToken, t.tokenExpires = token.AccessToken, token.RefreshToken, token.ExpiresAt()
	if t.onToken != nil {
		return t.onToken(token)
	}
	return nil
}

func (t *Trakt) authJSON(ctx context.Context, endpoint string, payload any, target any) error {
	status, err := t.authJSONStatus(ctx, endpoint, payload, target)
	if err == nil && status != http.StatusOK {
		err = fmt.Errorf("Trakt authentication returned status %d", status)
	}
	return err
}

func (t *Trakt) authJSONStatus(ctx context.Context, endpoint string, payload any, target any) (int, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, TraktAuthBaseURL+endpoint, bytes.NewReader(data))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := doRequest(t.httpClient, req)
	if err != nil {
		return 0, fmt.Errorf("Trakt authentication request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, resp.Body)
		return resp.StatusCode, nil
	}
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return resp.StatusCode, fmt.Errorf("failed to decode Trakt authentication response: %w", err)
	}
	return resp.StatusCode, nil
}
