package storage

import (
	"context"
	"fmt"

	"github.com/stsolovey/order_tracker/internal/models"
)

func (s *Storage) Upsert(ctx context.Context, order *models.Order) (*models.Order, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("Storage Upsert starting transaction: %w", err)
	}

	if _, err := s.UpsertOrder(ctx, tx, order); err != nil {
		if err := tx.Rollback(ctx); err != nil {
			return nil, fmt.Errorf("Storage Upsert order tx.Rollback(ctx): %w", err)
		}

		return nil, fmt.Errorf("Storage Upsert order: %w", err)
	}

	if _, err := s.UpsertDelivery(ctx, tx, &order.Delivery); err != nil {
		if err := tx.Rollback(ctx); err != nil {
			return nil, fmt.Errorf("Storage Upsert delivery tx.Rollback(ctx): %w", err)
		}

		return nil, fmt.Errorf("Storage Upsert delivery: %w", err)
	}

	if _, err := s.UpsertPayment(ctx, tx, &order.Payment); err != nil {
		if err := tx.Rollback(ctx); err != nil {
			return nil, fmt.Errorf("Storage Upsert payment tx.Rollback(ctx): %w", err)
		}

		return nil, fmt.Errorf("Storage Upsert payment: %w", err)
	}

	if _, err := s.UpsertItems(ctx, tx, &order.Items); err != nil {
		if err := tx.Rollback(ctx); err != nil {
			return nil, fmt.Errorf("Storage Upsert items tx.Rollback(ctx): %w", err)
		}

		return nil, fmt.Errorf("Storage Upsert items: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("Storage Upsert committing transaction: %w", err)
	}

	return order, nil
}

func (s *Storage) UpsertOrder(ctx context.Context, q Querier, order *models.Order) (*models.Order, error) {
	existsQuery := `SELECT EXISTS(SELECT 1 FROM orders WHERE order_uid = $1);`

	var exists bool

	var query string

	if err := q.QueryRow(ctx, existsQuery, order.OrderUID).Scan(&exists); err != nil {
		return nil, fmt.Errorf("Storage UpsertOrder check existence: %w", err)
	}

	if exists {
		query = `
            UPDATE orders
            SET track_number = $2, entry = $3, locale = $4, internal_signature = $5, 
                customer_id = $6, delivery_service = $7, shardkey = $8, sm_id = $9, 
                date_created = $10, oof_shard = $11
            WHERE order_uid = $1;
        `
	} else {
		query = `
            INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, 
                customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);
        `
	}

	if _, err := q.Exec(ctx, query, order.OrderUID, order.TrackNumber, order.Entry, order.Locale,
		order.InternalSignature, order.CustomerID, order.DeliveryService, order.Shardkey, order.SMID,
		order.DateCreated, order.OOFShard); err != nil {
		return nil, fmt.Errorf("Storage UpsertOrder q.Exec(...): %w", err)
	}

	return order, nil
}

func (s *Storage) UpsertDelivery(ctx context.Context, q Querier, delivery *models.Delivery) (*models.Delivery, error) {
	existsQuery := `SELECT EXISTS(SELECT 1 FROM delivery WHERE order_uid = $1);`

	var exists bool

	var query string

	if err := q.QueryRow(ctx, existsQuery, delivery.OrderUID).Scan(&exists); err != nil {
		return nil, fmt.Errorf("Storage UpsertDelivery check existence: %w", err)
	}

	if exists {
		query = `
			UPDATE delivery
			SET name = $2, phone = $3, zip = $4, city = $5, address = $6, 
				region = $7, email = $8
			WHERE order_uid = $1;
		`
	} else {
		query = `
			INSERT INTO delivery (order_uid, name, phone, zip, city, address, region, email)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
	
		`
	}

	if _, err := q.Exec(ctx, query, delivery.OrderUID, delivery.Name, delivery.Phone,
		delivery.Zip, delivery.City, delivery.Address, delivery.Region, delivery.Email); err != nil {
		return nil, fmt.Errorf("Storage UpsertDelivery q.Exec(...): %w", err)
	}

	return delivery, nil
}

func (s *Storage) UpsertPayment(ctx context.Context, q Querier, payment *models.Payment) (*models.Payment, error) {
	existsQuery := `SELECT EXISTS(SELECT 1 FROM payment WHERE order_uid = $1);`

	var exists bool

	var query string

	if err := q.QueryRow(ctx, existsQuery, payment.OrderUID).Scan(&exists); err != nil {
		return nil, fmt.Errorf("Storage UpsertPayment check existence: %w", err)
	}

	if exists {
		query = `
			UPDATE payment
			SET transaction = $2, request_id = $3, currency = $4, provider = $5, 
				amount = $6, payment_dt = $7, bank = $8, delivery_cost = $9, 
				goods_total = $10, custom_fee = $11
			WHERE order_uid = $1;
		`
	} else {
		query = `
			INSERT INTO payment (order_uid, transaction, request_id, currency, provider, 
				amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);
		`
	}

	if _, err := q.Exec(ctx, query, payment.OrderUID, payment.Transaction, payment.RequestID,
		payment.Currency, payment.Provider, payment.Amount, payment.PaymentDT, payment.Bank,
		payment.DeliveryCost, payment.GoodsTotal, payment.CustomFee); err != nil {
		return nil, fmt.Errorf("Storage UpsertPayment q.Exec(...): %w", err)
	}

	return payment, nil
}

func (s *Storage) UpsertItems(ctx context.Context, q Querier, items *[]models.Item) (*[]models.Item, error) {
	if len(*items) == 0 {
		return items, nil
	}

	orderUID := (*items)[0].OrderUID

	deleteQuery := `DELETE FROM items WHERE order_uid = $1;`
	if _, err := q.Exec(ctx, deleteQuery, orderUID); err != nil {
		return nil, fmt.Errorf("Storage UpsertItems delete existing: %w", err)
	}

	insertQuery := `
		INSERT INTO items (order_uid, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);
	`

	for _, item := range *items {
		if _, err := q.Exec(ctx, insertQuery, item.OrderUID, item.TrackNumber, item.Price, item.RID, item.Name,
			item.Sale, item.Size, item.TotalPrice, item.NMID, item.Brand, item.Status); err != nil {
			return nil, fmt.Errorf("Storage UpsertItems insert: %w", err)
		}
	}

	return items, nil
}
