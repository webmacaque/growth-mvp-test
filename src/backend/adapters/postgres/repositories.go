package postgres

import (
	"context"
	"errors"
	"time"

	"growth-mvp/backend/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IntegrationRepository struct {
	db *pgxpool.Pool
}

func NewIntegrationRepository(db *pgxpool.Pool) *IntegrationRepository {
	return &IntegrationRepository{db: db}
}

func (r *IntegrationRepository) Upsert(ctx context.Context, shopID int64, input domain.ConnectTelegramInput) (domain.TelegramIntegration, error) {
	const q = `
INSERT INTO telegram_integrations (shop_id, bot_token, chat_id, enabled, created_at, updated_at)
VALUES ($1, $2, $3, $4, NOW(), NOW())
ON CONFLICT (shop_id)
DO UPDATE SET bot_token = EXCLUDED.bot_token, chat_id = EXCLUDED.chat_id, enabled = EXCLUDED.enabled, updated_at = NOW()
RETURNING id, shop_id, bot_token, chat_id, enabled, created_at, updated_at`

	var out domain.TelegramIntegration
	err := r.db.QueryRow(ctx, q, shopID, input.BotToken, input.ChatID, input.Enabled).
		Scan(&out.ID, &out.ShopID, &out.BotToken, &out.ChatID, &out.Enabled, &out.CreatedAt, &out.UpdatedAt)
	return out, err
}

func (r *IntegrationRepository) GetByShopID(ctx context.Context, shopID int64) (domain.TelegramIntegration, bool, error) {
	const q = `SELECT id, shop_id, bot_token, chat_id, enabled, created_at, updated_at FROM telegram_integrations WHERE shop_id = $1`
	var out domain.TelegramIntegration
	err := r.db.QueryRow(ctx, q, shopID).Scan(
		&out.ID, &out.ShopID, &out.BotToken, &out.ChatID, &out.Enabled, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.TelegramIntegration{}, false, nil
		}
		return domain.TelegramIntegration{}, false, err
	}
	return out, true, nil
}

type OrderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, shopID int64, input domain.CreateOrderInput) (domain.Order, error) {
	const q = `
INSERT INTO orders (shop_id, number, total, customer_name, created_at)
VALUES ($1, $2, $3, $4, NOW())
RETURNING id, shop_id, number, total, customer_name, created_at`
	var out domain.Order
	err := r.db.QueryRow(ctx, q, shopID, input.Number, input.Total, input.CustomerName).
		Scan(&out.ID, &out.ShopID, &out.Number, &out.Total, &out.CustomerName, &out.CreatedAt)
	return out, err
}

type SendLogRepository struct {
	db *pgxpool.Pool
}

func NewSendLogRepository(db *pgxpool.Pool) *SendLogRepository {
	return &SendLogRepository{db: db}
}

func (r *SendLogRepository) Reserve(ctx context.Context, shopID, orderID int64, message string, reservedAt time.Time) (bool, error) {
	const q = `
INSERT INTO telegram_send_log (shop_id, order_id, message, status, error, sent_at)
VALUES ($1, $2, $3, 'FAILED', 'reserved', $4)
ON CONFLICT (shop_id, order_id) DO NOTHING`
	tag, err := r.db.Exec(ctx, q, shopID, orderID, message, reservedAt)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 1, nil
}

func (r *SendLogRepository) Finalize(ctx context.Context, shopID, orderID int64, status domain.TelegramSendStatus, errText *string, sentAt time.Time) error {
	const q = `
UPDATE telegram_send_log
SET status = $3, error = $4, sent_at = $5
WHERE shop_id = $1 AND order_id = $2`
	_, err := r.db.Exec(ctx, q, shopID, orderID, status, errText, sentAt)
	return err
}

func (r *SendLogRepository) GetStatusStats(ctx context.Context, shopID int64, since time.Time) (*time.Time, int64, int64, error) {
	const q = `
SELECT
  MAX(sent_at) FILTER (WHERE status = 'SENT') AS last_sent_at,
  COUNT(*) FILTER (WHERE status = 'SENT' AND sent_at >= $2) AS sent_count,
  COUNT(*) FILTER (WHERE status = 'FAILED' AND sent_at >= $2) AS failed_count
FROM telegram_send_log
WHERE shop_id = $1`
	var lastSentAt *time.Time
	var sentCount int64
	var failedCount int64
	err := r.db.QueryRow(ctx, q, shopID, since).Scan(&lastSentAt, &sentCount, &failedCount)
	return lastSentAt, sentCount, failedCount, err
}
