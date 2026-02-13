package domain

import "time"

type TelegramSendStatus string

const (
	TelegramSendStatusSent   TelegramSendStatus = "SENT"
	TelegramSendStatusFailed TelegramSendStatus = "FAILED"
)

type TelegramIntegration struct {
	ID        int64     `json:"id"`
	ShopID    int64     `json:"shopId"`
	BotToken  string    `json:"botToken"`
	ChatID    string    `json:"chatId"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Order struct {
	ID           int64     `json:"id"`
	ShopID       int64     `json:"shopId"`
	Number       string    `json:"number"`
	Total        float64   `json:"total"`
	CustomerName string    `json:"customerName"`
	CreatedAt    time.Time `json:"createdAt"`
}

type TelegramSendLog struct {
	ID      int64
	ShopID  int64
	OrderID int64
	Message string
	Status  TelegramSendStatus
	Error   *string
	SentAt  time.Time
}
