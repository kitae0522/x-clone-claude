package mediaclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type MediaStatus struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	MediaType string `json:"mediaType"`
	MimeType  string `json:"mimeType"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Size      int64  `json:"size"`
	URL       string `json:"url"`
	Error     string `json:"error"`
}

type Client interface {
	GetStatus(ctx context.Context, mediaID string) (*MediaStatus, error)
	Delete(ctx context.Context, mediaID string) error
}

type client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func New(baseURL, apiKey string) Client {
	return &client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *client) GetStatus(ctx context.Context, mediaID string) (*MediaStatus, error) {
	url := fmt.Sprintf("%s/internal/media/%s/status", c.baseURL, mediaID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("media-service request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("media-service returned status %d", resp.StatusCode)
	}

	var result struct {
		Success bool        `json:"success"`
		Data    MediaStatus `json:"data"`
		Error   string      `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if !result.Success {
		return nil, fmt.Errorf("media-service error: %s", result.Error)
	}

	return &result.Data, nil
}

func (c *client) Delete(ctx context.Context, mediaID string) error {
	url := fmt.Sprintf("%s/internal/media/%s", c.baseURL, mediaID)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("media-service request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("media-service returned status %d", resp.StatusCode)
	}
	return nil
}
