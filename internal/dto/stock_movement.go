package dto

type StockMovementResponse struct {
	ID              string  `json:"id"`
	ProductID       string  `json:"product_id"`
	ProductName     *string `json:"product_name,omitempty"`
	Operation       string  `json:"operation"`
	Quantity        int     `json:"quantity"`
	BeforeQuantity  int     `json:"before_quantity"`
	AfterQuantity   int     `json:"after_quantity"`
	Reason          *string `json:"reason,omitempty"`
	CreatedByUserID *string `json:"created_by_user_id,omitempty"`
	CreatedAt       string  `json:"created_at"`
}
