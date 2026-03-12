package dto

type AddCartItemRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type UpdateCartItemRequest struct {
	Quantity int `json:"quantity"`
}

type CartProductCategoryResponse struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type CartProductStrengthResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Level int16  `json:"level"`
}

type CartProductFlavorResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CartProductTagResponse struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type CartProductResponse struct {
	ID            string                       `json:"id"`
	Name          string                       `json:"name"`
	Description   *string                      `json:"description,omitempty"`
	Price         string                       `json:"price"`
	Category      CartProductCategoryResponse  `json:"category"`
	StockQuantity int                          `json:"stock_quantity"`
	Unit          string                       `json:"unit"`
	IsActive      bool                         `json:"is_active"`
	Strength      *CartProductStrengthResponse `json:"strength"`
	Flavors       []CartProductFlavorResponse  `json:"flavors"`
	Tags          []CartProductTagResponse     `json:"tags"`
}

type CartItemResponse struct {
	ID       string              `json:"id"`
	Quantity int                 `json:"quantity"`
	Product  CartProductResponse `json:"product"`
	Subtotal string              `json:"subtotal"`
}

type CartResponse struct {
	ID            string             `json:"id"`
	UserID        string             `json:"user_id"`
	Items         []CartItemResponse `json:"items"`
	TotalQuantity int                `json:"total_quantity"`
	TotalAmount   string             `json:"total_amount"`
}
