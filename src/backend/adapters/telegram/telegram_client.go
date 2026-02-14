package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	sendTimeout time.Duration
	httpClient  *http.Client
}

func NewClient(sendTimeout time.Duration) *Client {
	if sendTimeout <= 0 {
		sendTimeout = 5 * time.Second
	}

	return &Client{
		sendTimeout: sendTimeout,
		httpClient: &http.Client{
			Timeout: sendTimeout,
		},
	}
}

func (c *Client) SendMessage(ctx context.Context, botToken, chatID, text string) error {
	if botToken == "" || chatID == "" {
		return fmt.Errorf("botToken and chatID must be non-empty")
	}

	sendCtx, cancel := context.WithTimeout(ctx, c.sendTimeout)
	defer cancel()

	reqBody, err := json.Marshal(map[string]string{
		"chat_id": chatID,
		"text":    text,
	})

	if err != nil {
		return fmt.Errorf("marshal telegram payload: %w", err)
	}

	req, err := http.NewRequestWithContext(
		sendCtx,
		http.MethodPost,
		fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken),
		bytes.NewReader(reqBody),
	)

	if err != nil {
		return fmt.Errorf("create telegram request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)

	if err != nil {
		return fmt.Errorf("telegram sendMessage request failed: %w", err)
	}

	defer resp.Body.Close()

	var out struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return fmt.Errorf("decode telegram response: %w", err)
	}

	if resp.StatusCode >= 400 || !out.OK {
		if out.Description == "" {
			out.Description = "unknown telegram error"
		}
		return fmt.Errorf("telegram sendMessage failed (status=%d): %s", resp.StatusCode, out.Description)
	}

	return nil
}
