package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stsolovey/order_tracker/internal/models"
)

func (s *Storage) GetPayment(ctx context.Context, tx pgx.Tx, orderUID string) (*models.Payment, error) {
	var payment models.Payment

	var paymentDT int64

	query := `
        SELECT transaction, request_id, currency, provider, amount, payment_dt,
               bank, delivery_cost, goods_total, custom_fee
        FROM payment 
        WHERE order_uid = $1;
    `

	err := tx.QueryRow(ctx, query, orderUID).Scan(
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
