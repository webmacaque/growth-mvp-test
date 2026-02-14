package tests

import (
	"context"
	"fmt"
	"sync"
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
	nextID    int64
	listItems []domain.OrderListItem
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

func (f *MockOrderRepo) List(_ context.Context, _ int64, limit, offset int) ([]domain.OrderListItem, error) {
	if offset >= len(f.listItems) {
		return []domain.OrderListItem{}, nil
	}
	end := offset + limit
	if end > len(f.listItems) {
		end = len(f.listItems)
	}
	return append([]domain.OrderListItem(nil), f.listItems[offset:end]...), nil
}

type MockSendLogRepo struct {
	mu       sync.Mutex
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
	return fmt.Sprintf("%d:%d", shopID, orderID)
}

func (f *MockSendLogRepo) Reserve(_ context.Context, shopID, orderID int64, message string, reservedAt time.Time) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

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
	f.mu.Lock()
	defer f.mu.Unlock()

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
	mu    sync.Mutex
	calls int
	errs  []error
}

func (f *MockTelegramClient) SendMessage(_ context.Context, _, _, _ string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.calls++
	if len(f.errs) == 0 {
		return nil
	}

	err := f.errs[0]
	f.errs = f.errs[1:]
	return err
}

func (f *MockTelegramClient) Calls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

func waitForLogStatus(t *testing.T, repo *MockSendLogRepo, shopID, orderID int64, want domain.TelegramSendStatus, timeout time.Duration) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	k := key(shopID, orderID)
	for time.Now().Before(deadline) {
		repo.mu.Lock()
		log := repo.logs[k]
		repo.mu.Unlock()
		if log.Status == want {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	repo.mu.Lock()
	defer repo.mu.Unlock()
	t.Fatalf("expected log status %s, got %s", want, repo.logs[k].Status)
}

func waitForCalls(t *testing.T, client *MockTelegramClient, want int, timeout time.Duration) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if client.Calls() >= want {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("expected at least %d calls, got %d", want, client.Calls())
}

func TestCreateOrderEnabledIntegrationQueuesNotification(t *testing.T) {
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

	svc := domain.NewService(integrationRepo, orderRepo, sendLogRepo, telegramClient, 3)
	out, err := svc.CreateOrder(context.Background(), 1, domain.CreateOrderInput{
		Number:       "A-0001",
		Total:        100,
		CustomerName: "Anna",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.SendStatus != domain.SendStatusPending {
		t.Fatalf("expected queued status, got %s", out.SendStatus)
	}

	waitForCalls(t, telegramClient, 1, time.Second)
	waitForLogStatus(t, sendLogRepo, 1, out.Order.ID, domain.TelegramSendStatusSent, time.Second)
	if telegramClient.Calls() != 1 {
		t.Fatalf("expected one send call, got %d", telegramClient.Calls())
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
	svc := domain.NewService(integrationRepo, orderRepo, sendLogRepo, telegramClient, 3)

	sendLogRepo.Reserve(context.Background(), 1, 1, "msg", time.Now())

	out, err := svc.CreateOrder(context.Background(), 1, domain.CreateOrderInput{
		Number:       "A-0001",
		Total:        100,
		CustomerName: "Anna",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if telegramClient.Calls() != 0 {
		t.Fatalf("expected 0 send calls, got %d", telegramClient.Calls())
	}

	if out.SendStatus != domain.SendStatusSkipped {
		t.Fatalf("expected skipped status, got %s", out.SendStatus)
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
	sendErr := fmt.Errorf("telegram timeout")
	telegramClient := &MockTelegramClient{
		errs: []error{sendErr, sendErr, sendErr},
	}
	svc := domain.NewService(integrationRepo, orderRepo, sendLogRepo, telegramClient, 3)

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

	if out.SendStatus != domain.SendStatusPending {
		t.Fatalf("expected queued status, got %s", out.SendStatus)
	}

	waitForCalls(t, telegramClient, 3, 2*time.Second)
	waitForLogStatus(t, sendLogRepo, 1, out.Order.ID, domain.TelegramSendStatusFailed, 2*time.Second)
	if telegramClient.Calls() != 3 {
		t.Fatalf("expected 3 send attempts, got %d", telegramClient.Calls())
	}
}

func TestListOrdersPagination(t *testing.T) {
	now := time.Now()
	orderRepo := &MockOrderRepo{
		listItems: []domain.OrderListItem{
			{ID: 1, ShopID: 1, Number: "A-1", CreatedAt: now, SendStatus: domain.SendStatusPending},
			{ID: 2, ShopID: 1, Number: "A-2", CreatedAt: now.Add(-time.Minute), SendStatus: domain.SendStatusSent},
			{ID: 3, ShopID: 1, Number: "A-3", CreatedAt: now.Add(-2 * time.Minute), SendStatus: domain.SendStatusFailed},
		},
	}
	svc := domain.NewService(&MockIntegrationRepo{}, orderRepo, NewMockSendLogRepo(), &MockTelegramClient{}, 3)

	out, err := svc.ListOrders(context.Background(), 1, 2, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Limit != 2 || out.Offset != 0 {
		t.Fatalf("unexpected pagination: limit=%d offset=%d", out.Limit, out.Offset)
	}
	if len(out.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(out.Items))
	}
	if out.Items[0].SendStatus == "" || out.Items[1].SendStatus == "" {
		t.Fatal("expected sendStatus in list items")
	}
	if !out.HasMore {
		t.Fatal("expected hasMore=true")
	}
}

func TestListOrdersNormalizesInvalidPagination(t *testing.T) {
	now := time.Now()
	orderRepo := &MockOrderRepo{
		listItems: []domain.OrderListItem{
			{ID: 1, ShopID: 1, Number: "A-1", CreatedAt: now, SendStatus: domain.SendStatusPending},
		},
	}
	svc := domain.NewService(&MockIntegrationRepo{}, orderRepo, NewMockSendLogRepo(), &MockTelegramClient{}, 3)

	out, err := svc.ListOrders(context.Background(), 1, -10, -5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Limit != 20 || out.Offset != 0 {
		t.Fatalf("expected normalized pagination limit=20 offset=0, got limit=%d offset=%d", out.Limit, out.Offset)
	}
}
