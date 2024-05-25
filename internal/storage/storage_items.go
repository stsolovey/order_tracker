package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/stsolovey/order_tracker/internal/models"
)

func (s *Storage) GetItems(ctx context.Context, tx pgx.Tx, orderUID string) ([]models.Item, error) {
	var items []models.Item

	query := `
        SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
        FROM items
        WHERE order_uid = $1;
    `

	rows, err := tx.Query(ctx, query, orderUID)
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
