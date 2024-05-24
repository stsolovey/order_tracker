package models

type Delivery struct {
	DeliveryID int    `json:"deliveryId"`
	OrderUID   string `json:"orderUid"`
	Name       string `json:"name"`
	Phone      string `json:"phone"`
	Zip        string `json:"zip,omitempty"`
	City       string `json:"city"`
	Address    string `json:"address"`
	Region     string `json:"region,omitempty"`
	Email      string `json:"email,omitempty"`
}
