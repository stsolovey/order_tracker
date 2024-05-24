package models

import "time"

type Payment struct {
	PaymentID    int       `json:"paymentId"`
	OrderUID     string    `json:"orderUid"`
	Transaction  string    `json:"transaction"`
	RequestID    string    `json:"requestId,omitempty"`
	Currency     string    `json:"currency"`
	Provider     string    `json:"provider"`
	Amount       float64   `json:"amount"`
	PaymentDT    time.Time `json:"paymentDt"`
	Bank         string    `json:"bank,omitempty"`
	DeliveryCost float64   `json:"deliveryCost,omitempty"`
	GoodsTotal   float64   `json:"goodsTotal"`
	CustomFee    float64   `json:"customFee"`
}
