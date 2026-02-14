package domain

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrShopNotIntegrated = errors.New("telegram integration not found")
)

type Service struct {
	integrations IntegrationRepository
	orders       OrderRepository
	sendLogs     SendLogRepository
	telegram     TelegramClient

	retryMaxAttempts int
	retryBaseDelay   time.Duration
}

func NewService(
	integrations IntegrationRepository,
	orders OrderRepository,
	sendLogs SendLogRepository,
	telegram TelegramClient,
	retryMaxAttempts int,
) *Service {
	if retryMaxAttempts <= 0 {
		retryMaxAttempts = 3
	}

	return &Service{
		integrations:     integrations,
		orders:           orders,
		sendLogs:         sendLogs,
		telegram:         telegram,
		retryMaxAttempts: retryMaxAttempts,
		retryBaseDelay:   500 * time.Millisecond,
	}
}

func (s *Service) ConnectTelegram(ctx context.Context, shopID int64, input ConnectTelegramInput) (TelegramIntegration, error) {
	return s.integrations.Upsert(ctx, shopID, input)
}

func (s *Service) ListOrders(ctx context.Context, shopID int64, limit, offset int) (ListOrdersResult, error) {
	if limit <= 0 {
		limit = 20
	}

	if limit > 100 {
		limit = 100
	}

	if offset < 0 {
		offset = 0
	}

	rows, err := s.orders.List(ctx, shopID, limit+1, offset)

	if err != nil {
		return ListOrdersResult{}, err
	}

	hasMore := len(rows) > limit

	if hasMore {
		rows = rows[:limit]
	}

	return ListOrdersResult{
		Items:   rows,
		Limit:   limit,
		Offset:  offset,
		HasMore: hasMore,
	}, nil
}

func (s *Service) CreateOrder(ctx context.Context, shopID int64, input CreateOrderInput) (OrderSendResult, error) {
	order, err := s.orders.Create(ctx, shopID, input)

	if err != nil {
		return OrderSendResult{}, err
	}

	integration, found, err := s.integrations.GetByShopID(ctx, shopID)

	if err != nil {
		return OrderSendResult{}, err
	}

	if !found || !integration.Enabled {
		return OrderSendResult{
			Order:      order,
			SendStatus: SendStatusSkipped,
		}, nil
	}

	message := fmt.Sprintf("Новый заказ %s на сумму %.2f ₽, клиент %s", order.Number, order.Total, order.CustomerName)
	reservedAt := time.Now()
	reserved, err := s.sendLogs.Reserve(ctx, shopID, order.ID, message, reservedAt)

	if err != nil {
		return OrderSendResult{}, err
	}

	if !reserved {
		return OrderSendResult{
			Order:      order,
			SendStatus: SendStatusSkipped,
		}, nil
	}

	go s.trySendTelegram(shopID, order.ID, integration.BotToken, integration.ChatID, message)

	return OrderSendResult{
		Order:      order,
		SendStatus: SendStatusPending,
	}, nil
}

func (s *Service) trySendTelegram(shopID, orderID int64, botToken, chatID, message string) {
	var sendErr error

	for attempt := 1; attempt <= s.retryMaxAttempts; attempt++ {
		sendErr = s.telegram.SendMessage(context.Background(), botToken, chatID, message)

		if sendErr == nil {
			_ = s.sendLogs.Finalize(context.Background(), shopID, orderID, TelegramSendStatusSent, nil, time.Now())
			return
		}

		if attempt < s.retryMaxAttempts {
			time.Sleep(s.retryBaseDelay * time.Duration(attempt))
		}
	}

	errText := sendErr.Error()
	_ = s.sendLogs.Finalize(context.Background(), shopID, orderID, TelegramSendStatusFailed, &errText, time.Now())
}

func (s *Service) GetTelegramStatus(ctx context.Context, shopID int64) (TelegramStatus, error) {
	integration, found, err := s.integrations.GetByShopID(ctx, shopID)

	if err != nil {
		return TelegramStatus{}, err
	}

	if !found {
		return TelegramStatus{
			Enabled:      false,
			MaskedChatID: "",
			SentCount:    0,
			FailedCount:  0,
		}, nil
	}

	since := time.Now().AddDate(0, 0, -7)
	lastSentAt, sentCount, failedCount, err := s.sendLogs.GetStatusStats(ctx, shopID, since)

	if err != nil {
		return TelegramStatus{}, err
	}

	return TelegramStatus{
		Enabled:      integration.Enabled,
		MaskedChatID: integration.ChatID,
		LastSentAt:   lastSentAt,
		SentCount:    sentCount,
		FailedCount:  failedCount,
	}, nil
}
