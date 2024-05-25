package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/stsolovey/order_tracker/internal/models"
)

func (s *Storage) Get(ctx context.Context, orderUID string) (*models.Order, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("Storage Get s.beginTransaction(ctx): %w", err)
	}

	order, err := s.GetOrder(ctx, tx, orderUID)
	if err != nil {
		if err := tx.Rollback(ctx); err != nil {
			return nil, fmt.Errorf("Get GetOrder Rollback: %w", err)
		}

		return nil, fmt.Errorf("failed to fetch order: %w", err)
	}

	delivery, err := s.GetDelivery(ctx, tx, orderUID)
	if err != nil {
		if err := tx.Rollback(ctx); err != nil {
			return nil, fmt.Errorf("Get GetDelivery Rollback: %w", err)
		}

		return nil, fmt.Errorf("Get GetDelivery: %w", err)
	}

	order.Delivery = *delivery

	payment, err := s.GetPayment(ctx, tx, orderUID)
	if err != nil {
		if err := tx.Rollback(ctx); err != nil {
			return nil, fmt.Errorf("Get GetPayment Rollback: %w", err)
		}

		return nil, fmt.Errorf("failed to fetch payment: %w", err)
	}

	order.Payment = *payment

	items, err := s.GetItems(ctx, tx, orderUID)
	if err != nil {
		if err := tx.Rollback(ctx); err != nil {
			return nil, fmt.Errorf("Get GetItems Rollback: %w", err)
		}

		return nil, fmt.Errorf("failed to fetch items: %w", err)
	}

	order.Items = items

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return order, nil
}

func (s *Storage) GetOrder(ctx context.Context, q Querier, orderUID string) (*models.Order, error) {
	var order models.Order

	var dateCreated int64

	query := `
        SELECT order_uid, track_number, entry, locale, internal_signature, customer_id, 
               delivery_service, shardkey, sm_id, date_created, oof_shard
        FROM orders 
        WHERE order_uid = $1;
    `

	err := q.QueryRow(ctx, query, orderUID).Scan(
		&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale,
		&order.InternalSignature, &order.CustomerID, &order.DeliveryService,
		&order.Shardkey, &order.SMID, &dateCreated, &order.OOFShard,
	)
	if err != nil {
		return nil, fmt.Errorf("GetOrder failed: %w", err)
	}

	order.DateCreated = time.Unix(dateCreated, 0)

	return &order, nil
}

func (s *Storage) GetDelivery(ctx context.Context, q Querier, orderUID string) (*models.Delivery, error) {
	var delivery models.Delivery

	query := `
        SELECT name, phone, zip, city, address, region, email
        FROM delivery 
        WHERE order_uid = $1;
    `

	err := q.QueryRow(ctx, query, orderUID).Scan(
		&delivery.Name, &delivery.Phone, &delivery.Zip,
		&delivery.City, &delivery.Address, &delivery.Region, &delivery.Email,
	)
	if err != nil {
		return nil, fmt.Errorf("GetDelivery failed: %w", err)
	}

	return &delivery, nil
}

func (s *Storage) GetPayment(ctx context.Context, q Querier, orderUID string) (*models.Payment, error) {
	var payment models.Payment

	var paymentDT int64

	query := `
        SELECT transaction, request_id, currency, provider, amount, payment_dt,
               bank, delivery_cost, goods_total, custom_fee
        FROM payment 
        WHERE order_uid = $1;
    `

	err := q.QueryRow(ctx, query, orderUID).Scan(
		&payment.Transaction, &payment.RequestID, &payment.Currency,
		&payment.Provider, &payment.Amount, &paymentDT,
		&payment.Bank, &payment.DeliveryCost, &payment.GoodsTotal, &payment.CustomFee,
	)
	if err != nil {
		return nil, fmt.Errorf("GetPayment failed: %w", err)
	}

	payment.PaymentDT = time.Unix(paymentDT, 0)

	return &payment, nil
}

func (s *Storage) GetItems(ctx context.Context, q Querier, orderUID string) ([]models.Item, error) {
	var items []models.Item

	query := `
        SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
        FROM items
        WHERE order_uid = $1;
    `

	rows, err := q.Query(ctx, query, orderUID)
	if err != nil {
		return nil, fmt.Errorf("GetItems failed during query execution: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		var item models.Item

		if err := rows.Scan(&item.ChrtID, &item.TrackNumber, &item.Price, &item.RID,
			&item.Name, &item.Sale, &item.Size, &item.TotalPrice, &item.NMID,
			&item.Brand, &item.Status); err != nil {
			return nil, fmt.Errorf("GetItems failed during data scan: %w", err)
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetItems encountered errors during row processing: %w", err)
	}

	return items, nil
}
