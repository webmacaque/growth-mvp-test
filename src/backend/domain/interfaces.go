package domain

import (
	"context"
	"time"
)

type IntegrationRepository interface {
	Upsert(ctx context.Context, shopID int64, input ConnectTelegramInput) (TelegramIntegration, error)
	GetByShopID(ctx context.Context, shopID int64) (TelegramIntegration, bool, error)
}

type OrderRepository interface {
	Create(ctx context.Context, shopID int64, input CreateOrderInput) (Order, error)
	List(ctx context.Context, shopID int64, limit, offset int) ([]OrderListItem, error)
}

type SendLogRepository interface {
	Reserve(ctx context.Context, shopID, orderID int64, message string, reservedAt time.Time) (bool, error)
	Finalize(ctx context.Context, shopID, orderID int64, status TelegramSendStatus, errText *string, sentAt time.Time) error
	GetStatusStats(ctx context.Context, shopID int64, since time.Time) (lastSentAt *time.Time, sentCount, failedCount int64, err error)
}

type TelegramClient interface {
	SendMessage(ctx context.Context, botToken, chatID, text string) error
}
