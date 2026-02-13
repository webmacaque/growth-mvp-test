package domain

import "time"

const (
	SendStatusSent    = "sent"
	SendStatusFailed  = "failed"
	SendStatusSkipped = "skipped"
)

type ConnectTelegramInput struct {
	BotToken string `json:"botToken" binding:"required"`
	ChatID   string `json:"chatId" binding:"required"`
	Enabled  bool   `json:"enabled"`
}

type CreateOrderInput struct {
	Number       string  `json:"number" binding:"required"`
	Total        float64 `json:"total" binding:"required,gt=0"`
	CustomerName string  `json:"customerName" binding:"required"`
}

type OrderSendResult struct {
	Order      Order   `json:"order"`
	SendStatus string  `json:"sendStatus"`
	SendError  *string `json:"sendError,omitempty"`
}

type TelegramStatus struct {
	Enabled      bool       `json:"enabled"`
	MaskedChatID string     `json:"chatId"`
	LastSentAt   *time.Time `json:"lastSentAt"`
	SentCount    int64      `json:"sentCount7d"`
	FailedCount  int64      `json:"failedCount7d"`
}
