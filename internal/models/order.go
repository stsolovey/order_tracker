package models

import "time"

type Order struct {
	OrderUID          string    `json:"order_uid"`
	TrackNumber       string    `json:"track_number"`
	Entry             string    `json:"entry,omitempty"`
	Locale            string    `json:"locale"`
	InternalSignature string    `json:"internal_signature,omitempty"`
	CustomerID        string    `json:"customer_id"`
	DeliveryService   string    `json:"delivery_service"`
	Shardkey          string    `json:"shardkey,omitempty"`
	SMID              int       `json:"sm_id,omitempty"`
	DateCreated       time.Time `json:"date_created"`
	OOFShard          string    `json:"oof_shard,omitempty"`
}
