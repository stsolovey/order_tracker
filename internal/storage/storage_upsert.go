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
		return nil, fmt.Errorf("storage.go Upsert starting transaction: %w", err)
	}

	shouldRollback := true

	defer func() {
		if shouldRollback {
			if err := tx.Rollback(ctx); err != nil {
				s.log.Warn("Failed to rollback transaction", err)
			}
		}
	}()

	var orderReturning *models.Order

	if orderReturning, err = s.UpsertOrder(ctx, tx, order); err != nil {
		return nil, fmt.Errorf("storage.go Upsert order: %w", err)
	}

	if delivery, err := s.UpsertDelivery(ctx, tx, &order.Delivery); err != nil {
		orderReturning.Delivery = *delivery

		return nil, fmt.Errorf("storage.go Upsert delivery: %w", err)
	}

	if payment, err := s.UpsertPayment(ctx, tx, &order.Payment); err != nil {
		orderReturning.Payment = *payment

		return nil, fmt.Errorf("storage.go Upsert payment: %w", err)
	}

	if items, err := s.UpsertItems(ctx, tx, order.Items); err != nil {
		orderReturning.Items = *items

		return nil, fmt.Errorf("storage.go Upsert items: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("storage.go Upsert committing transaction: %w", err)
	}

	shouldRollback = false

	return orderReturning, nil
}

func (s *Storage) UpsertOrder(ctx context.Context, q Querier, order *models.Order) (*models.Order, error) {
	query := `
		INSERT INTO orders (
			order_uid, track_number, entry, locale, internal_signature, customer_id,
			delivery_service, shardkey, sm_id, date_created, oof_shard
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		) ON CONFLICT (order_uid) DO UPDATE SET
			track_number = EXCLUDED.track_number,
			entry = EXCLUDED.entry,
			locale = EXCLUDED.locale,
			internal_signature = EXCLUDED.internal_signature,
			customer_id = EXCLUDED.customer_id,
			delivery_service = EXCLUDED.delivery_service,
			shardkey = EXCLUDED.shardkey,
			sm_id = EXCLUDED.sm_id,
			date_created = EXCLUDED.date_created,
			oof_shard = EXCLUDED.oof_shard
		RETURNING 
			order_uid, track_number, entry, locale, internal_signature, customer_id,
			delivery_service, shardkey, sm_id, date_created, oof_shard;
	`

	var returningOrder models.Order

	err := q.QueryRow(ctx, query,
		order.OrderUID, order.TrackNumber, order.Entry, order.Locale,
		order.InternalSignature, order.CustomerID, order.DeliveryService, order.Shardkey,
		order.SMID, order.DateCreated, order.OOFShard,
	).Scan(
		&returningOrder.OrderUID, &returningOrder.TrackNumber, &returningOrder.Entry, &returningOrder.Locale,
		&returningOrder.InternalSignature, &returningOrder.CustomerID, &returningOrder.DeliveryService,
		&returningOrder.Shardkey, &returningOrder.SMID, &returningOrder.DateCreated, &returningOrder.OOFShard,
	)
	if err != nil {
		return nil, fmt.Errorf("storage.go UpsertOrder q.QueryRow(...): %w", err)
	}

	return &returningOrder, nil
}

func (s *Storage) UpsertDelivery(ctx context.Context, q Querier, delivery *models.Delivery) (*models.Delivery, error) {
	query := `
        INSERT INTO delivery (
            order_uid, name, phone, zip, city, address, region, email
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8
        ) ON CONFLICT (order_uid) DO UPDATE SET
            name = EXCLUDED.name,
            phone = EXCLUDED.phone,
            zip = EXCLUDED.zip,
            city = EXCLUDED.city,
            address = EXCLUDED.address,
            region = EXCLUDED.region,
            email = EXCLUDED.email
        RETURNING 
            order_uid, name, phone, zip, city, address, region, email;
    `

	var returningDelivery models.Delivery

	err := q.QueryRow(ctx, query,
		delivery.OrderUID, delivery.Name, delivery.Phone, delivery.Zip,
		delivery.City, delivery.Address, delivery.Region, delivery.Email,
	).Scan(
		&returningDelivery.OrderUID, &returningDelivery.Name, &returningDelivery.Phone,
		&returningDelivery.Zip, &returningDelivery.City, &returningDelivery.Address,
		&returningDelivery.Region, &returningDelivery.Email,
	)
	if err != nil {
		return nil, fmt.Errorf("storage.go UpsertDelivery q.QueryRow(...): %w", err)
	}

	return &returningDelivery, nil
}

func (s *Storage) UpsertPayment(ctx context.Context, q Querier, payment *models.Payment) (*models.Payment, error) {
	query := `
		INSERT INTO payment (
			order_uid, transaction, request_id, currency, provider, amount, payment_dt,
			bank, delivery_cost, goods_total, custom_fee
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		) ON CONFLICT (order_uid) DO UPDATE SET
			transaction = EXCLUDED.transaction,
			request_id = EXCLUDED.request_id,
			currency = EXCLUDED.currency,
			provider = EXCLUDED.provider,
			amount = EXCLUDED.amount,
			payment_dt = EXCLUDED.payment_dt,
			bank = EXCLUDED.bank,
			delivery_cost = EXCLUDED.delivery_cost,
			goods_total = EXCLUDED.goods_total,
			custom_fee = EXCLUDED.custom_fee
		RETURNING 
			order_uid, transaction, request_id, currency, provider, amount, payment_dt,
			bank, delivery_cost, goods_total, custom_fee;
	`

	var returnedPayment models.Payment

	err := q.QueryRow(ctx, query,
		payment.OrderUID, payment.Transaction, payment.RequestID, payment.Currency, payment.Provider,
		payment.Amount, payment.PaymentDT, payment.Bank, payment.DeliveryCost, payment.GoodsTotal,
		payment.CustomFee,
	).Scan(
		&returnedPayment.OrderUID, &returnedPayment.Transaction, &returnedPayment.RequestID,
		&returnedPayment.Currency, &returnedPayment.Provider, &returnedPayment.Amount,
		&returnedPayment.PaymentDT, &returnedPayment.Bank, &returnedPayment.DeliveryCost,
		&returnedPayment.GoodsTotal, &returnedPayment.CustomFee,
	)
	if err != nil {
		return nil, fmt.Errorf("storage.go UpsertPayment q.QueryRow(...): %w", err)
	}

	return &returnedPayment, nil
}

func (s *Storage) UpsertItems(ctx context.Context, q Querier, items []models.Item) (*[]models.Item, error) {
	if len(items) == 0 {
		return &items, nil
	}

	valueStrings := make([]string, 0, len(items))
	valueArgs := make([]any, 0, len(items)*12) //nolint:mnd

	for i, item := range items {
		valueStrings = append(valueStrings,
			fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
				i*12+1, i*12+2, i*12+3, i*12+4, i*12+5, i*12+6, i*12+7, i*12+8, i*12+9, i*12+10, i*12+11, i*12+12)) //nolint:mnd

		valueArgs = append(valueArgs, item.OrderUID, item.ChrtID, item.TrackNumber, item.Price, item.RID, item.Name,
			item.Sale, item.Size, item.TotalPrice, item.NMID, item.Brand, item.Status)
	}

	insertQuery := fmt.Sprintf(`
        INSERT INTO items (order_uid, chrt_id, track_number, price, 
			rid, name, sale, size, total_price, nm_id, brand, status)
        VALUES %s
        ON CONFLICT (chrt_id) DO UPDATE SET
            track_number = EXCLUDED.track_number,
            price = EXCLUDED.price,
            rid = EXCLUDED.rid,
            name = EXCLUDED.name,
            sale = EXCLUDED.sale,
            size = EXCLUDED.size,
            total_price = EXCLUDED.total_price,
            nm_id = EXCLUDED.nm_id,
            brand = EXCLUDED.brand,
            status = EXCLUDED.status
        RETURNING chrt_id, order_uid, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status;
    `, strings.Join(valueStrings, ","))

	rows, err := q.Query(ctx, insertQuery, valueArgs...)
	if err != nil {
		return nil, fmt.Errorf("storage.go UpsertItems batch insert: %w", err)
	}

	defer rows.Close()

	var returnedItems []models.Item

	for rows.Next() {
		var returnedItem models.Item
		if err := rows.Scan(&returnedItem.ChrtID, &returnedItem.OrderUID, &returnedItem.TrackNumber,
			&returnedItem.Price, &returnedItem.RID, &returnedItem.Name, &returnedItem.Sale,
			&returnedItem.Size, &returnedItem.TotalPrice, &returnedItem.NMID, &returnedItem.Brand,
			&returnedItem.Status); err != nil {
			return nil, fmt.Errorf("storage.go UpsertItems retrieving result: %w", err)
		}

		returnedItems = append(returnedItems, returnedItem)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("storage.go UpsertItems processing rows: %w", err)
	}

	return &returnedItems, nil
}
