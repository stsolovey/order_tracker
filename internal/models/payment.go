package models

import "time"

type Payment struct {
	PaymentID    int       `json:"payment_id"`
	OrderUID     string    `json:"order_uid"`
	Transaction  string    `json:"transaction"`
	RequestID    string    `json:"request_id,omitempty"`
	Currency     string    `json:"currency"`
	Provider     string    `json:"provider"`
	Amount       float64   `json:"amount"`
	PaymentDT    time.Time `json:"payment_dt"`
	Bank         string    `json:"bank,omitempty"`
	DeliveryCost float64   `json:"delivery_cost,omitempty"`
	GoodsTotal   float64   `json:"goods_total"`
	CustomFee    float64   `json:"custom_fee"`
}
