package models

type Item struct {
	ItemID      int     `json:"item_id"`
	OrderUID    string  `json:"order_uid"`
	TrackNumber string  `json:"track_number"`
	Price       float64 `json:"price"`
	RID         string  `json:"rid,omitempty"`
	Name        string  `json:"name"`
	Sale        int     `json:"sale,omitempty"`
	Size        string  `json:"size,omitempty"`
	TotalPrice  float64 `json:"total_price"`
	NMID        int     `json:"nm_id"`
	Brand       string  `json:"brand"`
	Status      int     `json:"status"`
}
