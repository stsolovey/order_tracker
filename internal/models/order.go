package models

import (
	"errors"
	"time"
)

var ErrOrderNotFound = errors.New("order not found")

type Order struct {
	OrderUID          string    `json:"orderUid"`
	TrackNumber       string    `json:"trackNumber"`
	Entry             string    `json:"entry,omitempty"`
	Locale            string    `json:"locale"`
	InternalSignature string    `json:"internalSignature,omitempty"`
	CustomerID        string    `json:"customerId"`
	DeliveryService   string    `json:"deliveryService"`
	Shardkey          string    `json:"shardkey,omitempty"`
	SMID              int       `json:"smId,omitempty"`
	DateCreated       time.Time `json:"dateCreated"`
	OOFShard          string    `json:"oofShard,omitempty"`
	Delivery          Delivery  `json:"delivery"`
	Payment           Payment   `json:"payment"`
	Items             []Item    `json:"items"`
}

type Delivery struct {
	OrderUID string `json:"orderUid"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Zip      string `json:"zip,omitempty"`
	City     string `json:"city"`
	Address  string `json:"address"`
	Region   string `json:"region,omitempty"`
	Email    string `json:"email,omitempty"`
}

type Payment struct {
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

type Item struct {
	OrderUID    string  `json:"orderUid"`
	ChrtID      int     `json:"chrtId"`
	TrackNumber string  `json:"trackNumber"`
	Price       float64 `json:"price"`
	RID         string  `json:"rid,omitempty"`
	Name        string  `json:"name"`
	Sale        int     `json:"sale,omitempty"`
	Size        string  `json:"size,omitempty"`
	TotalPrice  float64 `json:"totalPrice"`
	NMID        int     `json:"nmId"`
	Brand       string  `json:"brand"`
	Status      int     `json:"status"`
}
