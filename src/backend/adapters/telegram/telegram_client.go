package telegram

import "context"

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) SendMessage(_ context.Context, _, _, _ string) error {
	return nil
}
