package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"home-telemetry/agent/internal/types"
)

type Client struct {
	baseURL  string
	auth     string
	httpc    *http.Client
}

func New(baseURL, auth string) *Client {
	return &Client{
		baseURL: baseURL,
		auth:    auth,
		httpc: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) Send(payload types.IngestPayload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.baseURL+"/api/v1/ingest", bytes.NewReader(body))
	if err != nil {
		return err
	}
	if c.auth != "" {
		req.Header.Set("Authorization", "Bearer "+c.auth)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("ingest failed: %s", resp.Status)
	}
	return nil
}