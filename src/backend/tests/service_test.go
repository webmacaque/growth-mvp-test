package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"growth-mvp/backend/domain"
)

type MockIntegrationRepo struct {
	integration domain.TelegramIntegration
	found       bool
}

func (f *MockIntegrationRepo) Upsert(_ context.Context, shopID int64, input domain.ConnectTelegramInput) (domain.TelegramIntegration, error) {
	f.integration = domain.TelegramIntegration{
		ID:       1,
		ShopID:   shopID,
		BotToken: input.BotToken,
		ChatID:   input.ChatID,
		Enabled:  input.Enabled,
	}
	f.found = true
	return f.integration, nil
}

func (f *MockIntegrationRepo) GetByShopID(_ context.Context, _ int64) (domain.TelegramIntegration, bool, error) {
	return f.integration, f.found, nil
}

type MockOrderRepo struct {
	nextID int64
}

func (f *MockOrderRepo) Create(_ context.Context, shopID int64, input domain.CreateOrderInput) (domain.Order, error) {
	f.nextID++
	return domain.Order{
		ID:           f.nextID,
		ShopID:       shopID,
		Number:       input.Number,
		Total:        input.Total,
		CustomerName: input.CustomerName,
		CreatedAt:    time.Now(),
	}, nil
}

type MockSendLogRepo struct {
	reserved map[string]bool
	logs     map[string]domain.TelegramSendLog
}

func NewMockSendLogRepo() *MockSendLogRepo {
	return &MockSendLogRepo{
		reserved: map[string]bool{},
		logs:     map[string]domain.TelegramSendLog{},
	}
}

func key(shopID, orderID int64) string {
	return string(rune(shopID)) + ":" + string(rune(orderID))
}

func (f *MockSendLogRepo) Reserve(_ context.Context, shopID, orderID int64, message string, reservedAt time.Time) (bool, error) {
	k := key(shopID, orderID)
	if f.reserved[k] {
		return false, nil
	}
	f.reserved[k] = true
	f.logs[k] = domain.TelegramSendLog{
		ShopID:  shopID,
		OrderID: orderID,
		Message: message,
		Status:  domain.TelegramSendStatusFailed,
		SentAt:  reservedAt,
	}
	return true, nil
}

func (f *MockSendLogRepo) Finalize(_ context.Context, shopID, orderID int64, status domain.TelegramSendStatus, errText *string, sentAt time.Time) error {
	k := key(shopID, orderID)
	log := f.logs[k]
	log.Status = status
	log.Error = errText
	log.SentAt = sentAt
	f.logs[k] = log
	return nil
}

func (f *MockSendLogRepo) GetStatusStats(_ context.Context, _ int64, _ time.Time) (*time.Time, int64, int64, error) {
	return nil, 0, 0, nil
}

type MockTelegramClient struct {
	calls int
	err   error
}

func (f *MockTelegramClient) SendMessage(_ context.Context, _, _, _ string) error {
	f.calls++
	return f.err
}

func TestCreateOrderEnabledIntegrationWritesSentLog(t *testing.T) {
	integrationRepo := &MockIntegrationRepo{
		found: true,
		integration: domain.TelegramIntegration{
			ShopID:   1,
			BotToken: "token",
			ChatID:   "chat",
			Enabled:  true,
		},
	}

	orderRepo := &MockOrderRepo{}
	sendLogRepo := NewMockSendLogRepo()
	telegramClient := &MockTelegramClient{}

	svc := domain.NewService(integrationRepo, orderRepo, sendLogRepo, telegramClient)
	out, err := svc.CreateOrder(context.Background(), 1, domain.CreateOrderInput{
		Number:       "A-1",
		Total:        100,
		CustomerName: "Ann",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.SendStatus != "sent" {
		t.Fatalf("expected sent status, got %s", out.SendStatus)
	}

	if telegramClient.calls != 1 {
		t.Fatalf("expected telegram send call once, got %d", telegramClient.calls)
	}

	k := key(1, out.Order.ID)

	if sendLogRepo.logs[k].Status != domain.TelegramSendStatusSent {
		t.Fatalf("expected send log SENT, got %s", sendLogRepo.logs[k].Status)
	}
}

func TestDuplicateSendDoesNotSendAgain(t *testing.T) {
	integrationRepo := &MockIntegrationRepo{
		found: true,
		integration: domain.TelegramIntegration{
			ShopID:   1,
			BotToken: "token",
			ChatID:   "chat",
			Enabled:  true,
		},
	}
	orderRepo := &MockOrderRepo{}
	sendLogRepo := NewMockSendLogRepo()
	telegramClient := &MockTelegramClient{}
	svc := domain.NewService(integrationRepo, orderRepo, sendLogRepo, telegramClient)

	first, err := svc.CreateOrder(context.Background(), 1, domain.CreateOrderInput{
		Number:       "A-1",
		Total:        100,
		CustomerName: "Ann",
	})

	if err != nil {
		t.Fatalf("unexpected error on first call: %v", err)
	}

	reservedAgain, err := sendLogRepo.Reserve(context.Background(), 1, first.Order.ID, "same", time.Now())

	if err != nil {
		t.Fatalf("unexpected reserve error: %v", err)
	}

	if reservedAgain {
		t.Fatal("expected duplicate reserve to be rejected")
	}
}

func TestTelegramFailureDoesNotBreakOrderCreation(t *testing.T) {
	integrationRepo := &MockIntegrationRepo{
		found: true,
		integration: domain.TelegramIntegration{
			ShopID:   1,
			BotToken: "token",
			ChatID:   "chat",
			Enabled:  true,
		},
	}

	orderRepo := &MockOrderRepo{}
	sendLogRepo := NewMockSendLogRepo()
	telegramClient := &MockTelegramClient{err: errors.New("telegram timeout")}
	svc := domain.NewService(integrationRepo, orderRepo, sendLogRepo, telegramClient)

	out, err := svc.CreateOrder(context.Background(), 1, domain.CreateOrderInput{
		Number:       "A-0002",
		Total:        200,
		CustomerName: "Василий",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.Order.ID == 0 {
		t.Fatal("expected order to be created")
	}

	if out.SendStatus != "failed" {
		t.Fatalf("expected failed status, got %s", out.SendStatus)
	}

	k := key(1, out.Order.ID)

	if sendLogRepo.logs[k].Status != domain.TelegramSendStatusFailed {
		t.Fatalf("expected failed log, got %s", sendLogRepo.logs[k].Status)
	}
}
