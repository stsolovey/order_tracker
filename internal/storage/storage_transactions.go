package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/stsolovey/order_tracker/internal/models"
)

func (s *Storage) Get(ctx context.Context, orderUID string) (*models.Order, error) {
	txOptions := pgx.TxOptions{
		IsoLevel:       pgx.Serializable,
		AccessMode:     pgx.ReadWrite,
		DeferrableMode: pgx.NotDeferrable,
	}

	tx, err := s.db.BeginTx(ctx, txOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
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
