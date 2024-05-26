package storage

import (
	"context"
	"fmt"
	"strings"

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

	if err := q.QueryRow(ctx, existsQuery, order.OrderUID).Scan(&exists); err != nil {
		return nil, fmt.Errorf("Storage UpsertOrder check existence: %w", err)
	}

	var query string

	if exists {
		query = `
            UPDATE orders
            SET track_number = $2, entry = $3, locale = $4, internal_signature = $5, 
                customer_id = $6, delivery_service = $7, shardkey = $8, sm_id = $9, 
                date_created = $10, oof_shard = $11
            WHERE order_uid = $1
            RETURNING order_uid, track_number, entry, locale, internal_signature, customer_id, 
                      delivery_service, shardkey, sm_id, date_created, oof_shard;
        `
	} else {
		query = `
            INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, 
                customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
            RETURNING order_uid, track_number, entry, locale, internal_signature, customer_id, 
                      delivery_service, shardkey, sm_id, date_created, oof_shard;
        `
	}

	var returningOrder models.Order

	err := q.QueryRow(ctx, query, order.OrderUID, order.TrackNumber, order.Entry, order.Locale,
		order.InternalSignature, order.CustomerID, order.DeliveryService, order.Shardkey, order.SMID,
		order.DateCreated, order.OOFShard).Scan(
		&returningOrder.OrderUID, &returningOrder.TrackNumber, &returningOrder.Entry, &returningOrder.Locale,
		&returningOrder.InternalSignature, &returningOrder.CustomerID, &returningOrder.DeliveryService,
		&returningOrder.Shardkey, &returningOrder.SMID, &returningOrder.DateCreated, &returningOrder.OOFShard,
	)
	if err != nil {
		return nil, fmt.Errorf("Storage UpsertOrder q.Exec(...): %w", err)
	}

	return &returningOrder, nil
}

func (s *Storage) UpsertDelivery(ctx context.Context, q Querier, delivery *models.Delivery) (*models.Delivery, error) {
	existsQuery := `SELECT EXISTS(SELECT 1 FROM delivery WHERE order_uid = $1);`

	var exists bool

	if err := q.QueryRow(ctx, existsQuery, delivery.OrderUID).Scan(&exists); err != nil {
		return nil, fmt.Errorf("Storage UpsertDelivery check existence: %w", err)
	}

	var query string

	if exists {
		query = `
            UPDATE delivery
            SET name = $2, phone = $3, zip = $4, city = $5, address = $6, 
                region = $7, email = $8
            WHERE order_uid = $1
            RETURNING order_uid, name, phone, zip, city, address, region, email;
        `
	} else {
		query = `
            INSERT INTO delivery (order_uid, name, phone, zip, city, address, region, email)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
            RETURNING order_uid, name, phone, zip, city, address, region, email;
        `
	}

	var returningDelivery models.Delivery

	err := q.QueryRow(ctx, query, delivery.OrderUID, delivery.Name, delivery.Phone,
		delivery.Zip, delivery.City, delivery.Address, delivery.Region, delivery.Email).Scan(
		&returningDelivery.OrderUID, &returningDelivery.Name, &returningDelivery.Phone,
		&returningDelivery.Zip, &returningDelivery.City, &returningDelivery.Address,
		&returningDelivery.Region, &returningDelivery.Email,
	)
	if err != nil {
		return nil, fmt.Errorf("Storage UpsertDelivery q.QueryRow(...): %w", err)
	}

	return &returningDelivery, nil
}

func (s *Storage) UpsertPayment(ctx context.Context, q Querier, payment *models.Payment) (*models.Payment, error) {
	existsQuery := `SELECT EXISTS(SELECT 1 FROM payment WHERE order_uid = $1);`

	var exists bool

	if err := q.QueryRow(ctx, existsQuery, payment.OrderUID).Scan(&exists); err != nil {
		return nil, fmt.Errorf("Storage UpsertPayment check existence: %w", err)
	}

	var query string

	if exists {
		query = `
            UPDATE payment
            SET transaction = $2, request_id = $3, currency = $4, provider = $5, 
                amount = $6, payment_dt = $7, bank = $8, delivery_cost = $9, 
                goods_total = $10, custom_fee = $11
            WHERE order_uid = $1
            RETURNING order_uid, transaction, request_id, currency, provider, amount, payment_dt,
                      bank, delivery_cost, goods_total, custom_fee;
        `
	} else {
		query = `
            INSERT INTO payment (order_uid, transaction, request_id, currency, provider, 
                amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
            RETURNING order_uid, transaction, request_id, currency, provider, amount, payment_dt,
                      bank, delivery_cost, goods_total, custom_fee;
        `
	}

	var returnedPayment models.Payment

	err := q.QueryRow(ctx, query, payment.OrderUID, payment.Transaction, payment.RequestID,
		payment.Currency, payment.Provider, payment.Amount, payment.PaymentDT, payment.Bank,
		payment.DeliveryCost, payment.GoodsTotal, payment.CustomFee).Scan(
		&returnedPayment.OrderUID, &returnedPayment.Transaction, &returnedPayment.RequestID,
		&returnedPayment.Currency, &returnedPayment.Provider, &returnedPayment.Amount,
		&returnedPayment.PaymentDT, &returnedPayment.Bank, &returnedPayment.DeliveryCost,
		&returnedPayment.GoodsTotal, &returnedPayment.CustomFee,
	)
	if err != nil {
		return nil, fmt.Errorf("Storage UpsertPayment query execution: %w", err)
	}

	return &returnedPayment, nil
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

	var valueStrings []string
	var valueArgs []interface{}
	for i, item := range *items {
		valueStrings = append(valueStrings,
			fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
				i*11+1, i*11+2, i*11+3, i*11+4, i*11+5, i*11+6, i*11+7, i*11+8, i*11+9, i*11+10, i*11+11))
		valueArgs = append(valueArgs, item.OrderUID, item.TrackNumber, item.Price, item.RID, item.Name,
			item.Sale, item.Size, item.TotalPrice, item.NMID, item.Brand, item.Status)
	}

	insertQuery := fmt.Sprintf(`
        INSERT INTO items (order_uid, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
        VALUES %s
        RETURNING chrt_id, order_uid, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status;
    `, strings.Join(valueStrings, ","))

	rows, err := q.Query(ctx, insertQuery, valueArgs...)
	if err != nil {
		return nil, fmt.Errorf("Storage UpsertItems batch insert: %w", err)
	}
	defer rows.Close()

	var returnedItems []models.Item
	for rows.Next() {
		var returnedItem models.Item
		if err := rows.Scan(&returnedItem.ChrtID, &returnedItem.OrderUID, &returnedItem.TrackNumber,
			&returnedItem.Price, &returnedItem.RID, &returnedItem.Name, &returnedItem.Sale,
			&returnedItem.Size, &returnedItem.TotalPrice, &returnedItem.NMID, &returnedItem.Brand,
			&returnedItem.Status); err != nil {
			return nil, fmt.Errorf("Storage UpsertItems retrieving result: %w", err)
		}
		returnedItems = append(returnedItems, returnedItem)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Storage UpsertItems processing rows: %w", err)
	}

	return &returnedItems, nil
}
