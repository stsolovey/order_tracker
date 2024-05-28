package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stsolovey/order_tracker/internal/models"
)

func (s *Storage) GetAll(ctx context.Context) ([]models.Order, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("Storage GetAll(...) beginTransaction(...): %w", err)
	}

	orders, deliveries, payments, items, err := s.fetchAllData(ctx, tx)
	if err != nil {
		if err = tx.Rollback(ctx); err != nil {
			return nil, fmt.Errorf("Storage GetAll(...) s.fetchAllData tx.Rollback(...): %w", err)
		}

		return nil, fmt.Errorf("Storage GetAll(...) s.fetchAllData(...): %w", err)
	}

	orders = s.associateData(orders, deliveries, payments, items)

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("Storage GetAll(...) tx.Commit(...): %w", err)
	}

	return orders, nil
}

func (s *Storage) fetchAllData(ctx context.Context, tx pgx.Tx) (
	[]models.Order,
	[]models.Delivery,
	[]models.Payment,
	[]models.Item,
	error,
) {
	orders, err := s.GetOrders(ctx, tx)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("Storage GetAll(...) GetOrders(...): %w", err)
	}

	deliveries, err := s.GetDeliveries(ctx, tx)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("Storage GetAll(...) GetDeliveries(...): %w", err)
	}

	payments, err := s.GetPayments(ctx, tx)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("Storage GetAll(...) GetPayments(...): %w", err)
	}

	items, err := s.GetItemsAll(ctx, tx)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("Storage GetAll(...) GetItemsAll(...): %w", err)
	}

	return orders, deliveries, payments, items, nil
}

func (s *Storage) associateData(
	orders []models.Order,
	deliveries []models.Delivery,
	payments []models.Payment,
	items []models.Item,
) []models.Order {
	deliveryMap := make(map[string]models.Delivery)
	for _, delivery := range deliveries {
		deliveryMap[delivery.OrderUID] = delivery
	}

	paymentMap := make(map[string]models.Payment)
	for _, payment := range payments {
		paymentMap[payment.OrderUID] = payment
	}

	itemsMap := make(map[string][]models.Item)
	for _, item := range items {
		itemsMap[item.OrderUID] = append(itemsMap[item.OrderUID], item)
	}

	for i, order := range orders {
		if delivery, ok := deliveryMap[order.OrderUID]; ok {
			orders[i].Delivery = delivery
		}

		if payment, ok := paymentMap[order.OrderUID]; ok {
			orders[i].Payment = payment
		}

		if items, ok := itemsMap[order.OrderUID]; ok {
			orders[i].Items = items
		}
	}

	return orders
}

func (s *Storage) GetOrders(ctx context.Context, q Querier) ([]models.Order, error) {
	query := `
        SELECT order_uid, track_number, entry, locale, internal_signature, customer_id, 
               delivery_service, shardkey, sm_id, date_created, oof_shard
        FROM orders;
    `

	rows, err := q.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("Storage GetOrders(...) q.Query(...): %w", err)
	}

	defer rows.Close()

	var orders []models.Order

	for rows.Next() {
		var order models.Order

		var dateCreated int64

		if err := rows.Scan(
			&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale,
			&order.InternalSignature, &order.CustomerID, &order.DeliveryService,
			&order.Shardkey, &order.SMID, &dateCreated, &order.OOFShard,
		); err != nil {
			return nil, fmt.Errorf("Storage GetOrders(...) rows.Scan(...): %w", err)
		}

		order.DateCreated = time.Unix(dateCreated, 0)

		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Storage GetOrders(...) rows.Err(...): %w", err)
	}

	return orders, nil
}

func (s *Storage) GetDeliveries(ctx context.Context, q Querier) ([]models.Delivery, error) {
	query := `
        SELECT order_uid, name, phone, zip, city, address, region, email
        FROM delivery;
    `

	rows, err := q.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("Storage GetDeliveries(...) q.Query(...): %w", err)
	}

	defer rows.Close()

	var deliveries []models.Delivery

	for rows.Next() {
		var delivery models.Delivery

		if err := rows.Scan(
			&delivery.OrderUID, &delivery.Name, &delivery.Phone, &delivery.Zip,
			&delivery.City, &delivery.Address, &delivery.Region, &delivery.Email,
		); err != nil {
			return nil, fmt.Errorf("Storage GetDeliveries(...) rows.Scan(...): %w", err)
		}

		deliveries = append(deliveries, delivery)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Storage GetDeliveries(...) rows.Err(...): %w", err)
	}

	return deliveries, nil
}

func (s *Storage) GetItemsAll(ctx context.Context, q Querier) ([]models.Item, error) {
	query := `
        SELECT order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
        FROM items;
    `

	rows, err := q.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("Storage GetItemsAll(...) q.Query(...): %w", err)
	}

	defer rows.Close()

	var items []models.Item

	for rows.Next() {
		var item models.Item

		if err := rows.Scan(
			&item.OrderUID, &item.ChrtID, &item.TrackNumber, &item.Price, &item.RID,
			&item.Name, &item.Sale, &item.Size, &item.TotalPrice, &item.NMID,
			&item.Brand, &item.Status,
		); err != nil {
			return nil, fmt.Errorf("Storage GetItemsAll(...) rows.Scan(...): %w", err)
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Storage GetItemsAll(...) rows.Err(...): %w", err)
	}

	return items, nil
}

func (s *Storage) GetPayments(ctx context.Context, q Querier) ([]models.Payment, error) {
	query := `
        SELECT order_uid, transaction, request_id, currency, provider, amount, payment_dt,
               bank, delivery_cost, goods_total, custom_fee
        FROM payment;
    `

	rows, err := q.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("Storage GetPayments(...) q.Query(...): %w", err)
	}

	defer rows.Close()

	var payments []models.Payment

	for rows.Next() {
		var payment models.Payment

		var paymentDT int64

		if err := rows.Scan(
			&payment.OrderUID, &payment.Transaction, &payment.RequestID, &payment.Currency,
			&payment.Provider, &payment.Amount, &paymentDT, &payment.Bank,
			&payment.DeliveryCost, &payment.GoodsTotal, &payment.CustomFee,
		); err != nil {
			return nil, fmt.Errorf("Storage GetPayments(...) rows.Scan(...): %w", err)
		}

		payment.PaymentDT = time.Unix(paymentDT, 0)

		payments = append(payments, payment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Storage GetPayments(...) rows.Err(...): %w", err)
	}

	return payments, nil
}
