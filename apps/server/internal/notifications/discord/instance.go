package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mahcks/blockbusterr/internal/global"
)

type Service interface {
	SendDiscordEmbed(webhookURL string, embed Embed) error
}

type discordService struct {
	gctx global.Context
}

// EmbedField represents a field in the embedded message
type EmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"` // Inline allows for side-by-side fields
}

// Embed represents the embed structure for the message
type Embed struct {
	Title       string       `json:"title,omitempty"`
	Description string       `json:"description,omitempty"`
	Color       int          `json:"color,omitempty"`  // Decimal value of the embed's color
	Fields      []EmbedField `json:"fields,omitempty"` // Array of fields to display in the embed
}

// Payload represents the JSON structure you will send to the webhook
type Payload struct {
	Content string  `json:"content,omitempty"` // Optional message content outside of embed
	Embeds  []Embed `json:"embeds"`            // Embeds is an array of Embed structs
}

func (d *discordService) SendDiscordEmbed(webhookURL string, embed Embed) error {
	// Create the payload with the embed
	payload := Payload{
		Embeds: []Embed{embed},
	}

	// Convert the payload into JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	// Send the POST request
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
