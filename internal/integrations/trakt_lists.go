package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

func (t *Trakt) GetListItems(ctx context.Context, owner, listID string, watchlist bool, limit int) (TraktListItems, error) {
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
	pages, err := t.doRequestPaginated(ctx, endpoint, limit)
	if err != nil {
		return TraktListItems{}, err
	}
	result := TraktListItems{}
	for _, page := range pages {
		var items []struct {
			Type  string `json:"type"`
			Movie *Movie `json:"movie"`
			Show  *Show  `json:"show"`
		}
		if err := json.Unmarshal(page, &items); err != nil {
			return TraktListItems{}, fmt.Errorf("failed to decode Trakt list: %w", err)
		}
		for _, item := range items {
			if item.Type == "movie" && item.Movie != nil {
				result.Movies = append(result.Movies, *item.Movie)
			} else if item.Type == "show" && item.Show != nil {
				result.Shows = append(result.Shows, *item.Show)
			}
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
	return token, status, t.storeToken(token)
}

func (t *Trakt) refreshAccessToken(ctx context.Context) error {
	t.tokenMu.Lock()
	defer t.tokenMu.Unlock()
	if time.Now().Unix() < t.tokenExpires-60 {
		return nil
	}
	var token TraktToken
	status, err := t.authJSONStatus(ctx, "/oauth/token", map[string]string{"refresh_token": t.refreshToken, "client_id": t.clientID, "client_secret": t.clientSecret, "grant_type": "refresh_token", "redirect_uri": "urn:ietf:wg:oauth:2.0:oob"}, &token)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("Trakt token refresh returned status %d", status)
	}
	if token.CreatedAt == 0 {
		token.CreatedAt = time.Now().Unix()
	}
	return t.storeToken(token)
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
	resp, err := t.httpClient.Do(req)
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
