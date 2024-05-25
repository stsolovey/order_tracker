package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/stsolovey/order_tracker/internal/models"
)

func (s *Storage) GetDelivery(ctx context.Context, tx pgx.Tx, orderUID string) (*models.Delivery, error) {
	var delivery models.Delivery

	query := `
        SELECT name, phone, zip, city, address, region, email
        FROM delivery 
        WHERE order_uid = $1;
    `

	err := tx.QueryRow(ctx, query, orderUID).Scan(
		&delivery.Name, &delivery.Phone, &delivery.Zip,
		&delivery.City, &delivery.Address, &delivery.Region, &delivery.Email,
	)
	if err != nil {
		return nil, fmt.Errorf("GetDelivery failed: %w", err)
	}

	return &delivery, nil
}
