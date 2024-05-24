package models

import "time"

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
}
