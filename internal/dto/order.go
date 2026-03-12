package dto

type OrderItemResponse struct {
	ID              string `json:"id"`
	ProductID       string `json:"product_id"`
	ProductNameSnap string `json:"product_name_snapshot"`
	UnitPriceSnap   string `json:"unit_price_snapshot"`
	Quantity        int    `json:"quantity"`
	Subtotal        string `json:"subtotal"`
	CreatedAt       string `json:"created_at"`
}

type OrderResponse struct {
	ID          string              `json:"id"`
	UserID      string              `json:"user_id"`
	Status      string              `json:"status"`
	TotalAmount string              `json:"total_amount"`
	Items       []OrderItemResponse `json:"items"`
	CreatedAt   string              `json:"created_at"`
	UpdatedAt   string              `json:"updated_at"`
}

type CreateOrderResponse struct {
	ID          string              `json:"id"`
	Status      string              `json:"status"`
	TotalAmount string              `json:"total_amount"`
	Items       []OrderItemResponse `json:"items"`
	CreatedAt   string              `json:"created_at"`
}

type OrderUserSummary struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"`
}

type AdminOrderResponse struct {
	ID          string              `json:"id"`
	Status      string              `json:"status"`
	TotalAmount string              `json:"total_amount"`
	Items       []OrderItemResponse `json:"items"`
	User        OrderUserSummary    `json:"user"`
	CreatedAt   string              `json:"created_at"`
	UpdatedAt   string              `json:"updated_at"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status"`
}
