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
}

func NewService(
	integrations IntegrationRepository,
	orders OrderRepository,
	sendLogs SendLogRepository,
	telegram TelegramClient,
) *Service {
	return &Service{
		integrations: integrations,
		orders:       orders,
		sendLogs:     sendLogs,
		telegram:     telegram,
	}
}

func (s *Service) ConnectTelegram(ctx context.Context, shopID int64, input ConnectTelegramInput) (TelegramIntegration, error) {
	return s.integrations.Upsert(ctx, shopID, input)
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
			SendStatus: "skipped",
		}, nil
	}

	message := fmt.Sprintf("Заказ %s на сумму %.2f ₽, покупатель: %s", order.Number, order.Total, order.CustomerName)
	reservedAt := time.Now()
	reserved, err := s.sendLogs.Reserve(ctx, shopID, order.ID, message, reservedAt)

	if err != nil {
		return OrderSendResult{}, err
	}

	if !reserved {
		return OrderSendResult{
			Order:      order,
			SendStatus: "skipped",
		}, nil
	}

	if err := s.telegram.SendMessage(ctx, integration.BotToken, integration.ChatID, message); err != nil {
		errText := err.Error()

		if finalizeErr := s.sendLogs.Finalize(ctx, shopID, order.ID, TelegramSendStatusFailed, &errText, time.Now()); finalizeErr != nil {
			return OrderSendResult{}, finalizeErr
		}

		return OrderSendResult{
			Order:      order,
			SendStatus: "failed",
			SendError:  &errText,
		}, nil
	}

	if err := s.sendLogs.Finalize(ctx, shopID, order.ID, TelegramSendStatusSent, nil, time.Now()); err != nil {
		return OrderSendResult{}, err
	}

	return OrderSendResult{
		Order:      order,
		SendStatus: "sent",
	}, nil
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
