package models

// Product struct
type Product struct {
	ID          string  `json:"id,omitempty" bson:"_id,omitempty"`
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	Image       string  `json:"image"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Show        bool    `json:"show"`
}

type ProductResponse struct {
	Total int       `json:"total"`
	Items []Product `json:"items"`
}
