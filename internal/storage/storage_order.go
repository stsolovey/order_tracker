package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stsolovey/order_tracker/internal/models"
)

func (s *Storage) GetOrder(ctx context.Context, tx pgx.Tx, orderUID string) (*models.Order, error) {
	var order models.Order

	var dateCreated int64

	query := `
        SELECT order_uid, track_number, entry, locale, internal_signature, customer_id, 
               delivery_service, shardkey, sm_id, date_created, oof_shard
        FROM orders 
        WHERE order_uid = $1;
    `

	err := tx.QueryRow(ctx, query, orderUID).Scan(
		&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale,
		&order.InternalSignature, &order.CustomerID, &order.DeliveryService,
		&order.Shardkey, &order.SMID, &dateCreated, &order.OOFShard,
	)
	if err != nil {
		return nil, fmt.Errorf("GetOrder failed: %w", err)
	}

	order.DateCreated = time.Unix(dateCreated, 0) // Convert UNIX timestamp to time.Time

	return &order, nil
}
