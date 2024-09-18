package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mahcks/blockbusterr/internal/helpers"
	"github.com/mahcks/blockbusterr/internal/helpers/trakt"
	"github.com/mahcks/blockbusterr/pkg/structures"
)

type DiscordNotification struct {
	helpers    *helpers.Helpers
	WebhookURL string
	Client     *http.Client
	Username   string
}

func NewDiscordNotification(helpers *helpers.Helpers, webhookURL, username string) *DiscordNotification {
	return &DiscordNotification{
		helpers:    helpers,
		WebhookURL: webhookURL,
		Client:     &http.Client{Timeout: 10 * time.Second},
		Username:   username,
	}
}

type DiscordEmbed struct {
	Title     string `json:"title"`
	Thumbnail *struct {
		URL string `json:"url"`
		W   int    `json:"width"`
		H   int    `json:"height"`
	} `json:"thumbnail,omitempty"`
	Description string `json:"description"`
	Color       int    `json:"color,omitempty"`
	Author      *struct {
		Name string `json:"name,omitempty"`
	} `json:"author,omitempty"`
	Fields []DiscordEmbedField `json:"fields,omitempty"`
	Footer *struct {
		Text string `json:"text,omitempty"`
	} `json:"footer,omitempty"`
	Timestamp string `json:"timestamp,omitempty"` // ISO8601 timestamp
}

type DiscordEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type DiscordWebhookPayload struct {
	Username string         `json:"username,omitempty"`
	Embeds   []DiscordEmbed `json:"embeds"`
}

// CreateEmbed generates a Discord embed payload based on the movie information
func (d *DiscordNotification) CreateEmbed(title, poster, description string, year int, genre string, rating float64) DiscordEmbed {
	return DiscordEmbed{
		Title: title,
		Thumbnail: &struct {
			URL string "json:\"url\""
			W   int    "json:\"width\""
			H   int    "json:\"height\""
		}{
			URL: poster,
		},
		Description: description,
		Color:       0x00ff00, // Green color
		Author: &struct {
			Name string "json:\"name,omitempty\""
		}{
			Name: "blockbusterr",
		},
		Fields: []DiscordEmbedField{
			{Name: "Year", Value: fmt.Sprintf("%d", year), Inline: true},
			{Name: "Genre", Value: genre, Inline: true},
			{Name: "Rating", Value: fmt.Sprintf("%.1f", rating), Inline: true},
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

// SendNotification sends the Discord notification with the generated payload
func (d *DiscordNotification) SendNotification(notificationType structures.NotificationType, payload json.RawMessage) error {
	switch notificationType {
	case structures.MOVIEADDEDALERT:
		var movie trakt.Movie
		err := json.Unmarshal(payload, &movie)
		if err != nil {
			return fmt.Errorf("failed to unmarshal movie payload: %w", err)
		}

		omdbMedia, err := d.helpers.OMDb.GetMedia(context.Background(), movie.IDs.IMDB)
		if err != nil {
			return fmt.Errorf("failed to fetch movie details from OMDb: %w", err)
		}

		embed := d.CreateEmbed(
			movie.Title,
			omdbMedia.Poster,
			movie.Overview,
			movie.Year,
			strings.Join(movie.Genres, ", "),
			movie.Rating,
		)

		discordPayload := DiscordWebhookPayload{
			Username: d.Username,
			Embeds:   []DiscordEmbed{embed},
		}

		return d.sendDiscordPayload(discordPayload)

	case structures.SHOWADDEDALERT:
		var show trakt.Show
		err := json.Unmarshal(payload, &show)
		if err != nil {
			return fmt.Errorf("failed to unmarshal show payload: %w", err)
		}

		omdbMedia, err := d.helpers.OMDb.GetMedia(context.Background(), show.IDs.IMDB)
		if err != nil {
			return fmt.Errorf("failed to fetch movie details from OMDb: %w", err)
		}

		embed := d.CreateEmbed(
			show.Title,
			omdbMedia.Poster,
			show.Overview,
			show.Year,
			strings.Join(show.Genres, ", "),
			show.Rating,
		)

		discordPayload := DiscordWebhookPayload{
			Username: d.Username,
			Embeds:   []DiscordEmbed{embed},
		}

		return d.sendDiscordPayload(discordPayload)
	}

	return nil
}

func (d *DiscordNotification) sendDiscordPayload(payload DiscordWebhookPayload) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal discord embed payload: %w", err)
	}

	req, err := http.NewRequest("POST", d.WebhookURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create http request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := d.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send discord notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("discord webhook returned status code %d", resp.StatusCode)
	}

	return nil
}
