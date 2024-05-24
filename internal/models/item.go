package models

type Item struct {
	ItemID      int     `json:"itemId"`
	OrderUID    string  `json:"orderUid"`
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
