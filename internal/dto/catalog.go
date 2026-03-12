package dto

type CategoryResponse struct {
	ID          string  `json:"id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	IsActive    bool    `json:"is_active"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type CreateCategoryRequest struct {
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	IsActive    *bool   `json:"is_active"`
}

type UpdateCategoryRequest struct {
	Code        *string `json:"code"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
	IsActive    *bool   `json:"is_active"`
}

type FlavorResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	IsActive    bool    `json:"is_active"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type CreateFlavorRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	IsActive    *bool   `json:"is_active"`
}

type UpdateFlavorRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	IsActive    *bool   `json:"is_active"`
}

type StrengthResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Level       int16   `json:"level"`
	Description *string `json:"description,omitempty"`
	IsActive    bool    `json:"is_active"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type CreateStrengthRequest struct {
	Name        string  `json:"name"`
	Level       int16   `json:"level"`
	Description *string `json:"description"`
	IsActive    *bool   `json:"is_active"`
}

type UpdateStrengthRequest struct {
	Name        *string `json:"name"`
	Level       *int16  `json:"level"`
	Description *string `json:"description"`
	IsActive    *bool   `json:"is_active"`
}

type ProductCategoryInfo struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type ProductStrengthInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Level int16  `json:"level"`
}

type ProductFlavorInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ProductTagInfo struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type ProductResponse struct {
	ID            string               `json:"id"`
	Name          string               `json:"name"`
	Description   *string              `json:"description,omitempty"`
	Price         string               `json:"price"`
	StockQuantity int                  `json:"stock_quantity"`
	Unit          string               `json:"unit"`
	IsActive      bool                 `json:"is_active"`
	Category      ProductCategoryInfo  `json:"category"`
	Strength      *ProductStrengthInfo `json:"strength"`
	Flavors       []ProductFlavorInfo  `json:"flavors"`
	Tags          []ProductTagInfo     `json:"tags"`
	CreatedAt     string               `json:"created_at"`
	UpdatedAt     string               `json:"updated_at"`
}

type CreateProductRequest struct {
	CategoryID    string     `json:"category_id"`
	Name          string     `json:"name"`
	Description   *string    `json:"description"`
	Price         PriceInput `json:"price"`
	StockQuantity int        `json:"stock_quantity"`
	Unit          string     `json:"unit"`
	IsActive      *bool      `json:"is_active"`
	StrengthID    *string    `json:"strength_id"`
	FlavorIDs     []string   `json:"flavor_ids"`
}

type UpdateProductRequest struct {
	CategoryID    *string     `json:"category_id"`
	Name          *string     `json:"name"`
	Description   *string     `json:"description"`
	Price         *PriceInput `json:"price"`
	StockQuantity *int        `json:"stock_quantity"`
	Unit          *string     `json:"unit"`
	IsActive      *bool       `json:"is_active"`
	StrengthID    *string     `json:"strength_id"`
	FlavorIDs     *[]string   `json:"flavor_ids"`
}

type UpdateStockRequest struct {
	Operation string  `json:"operation"`
	Quantity  int     `json:"quantity"`
	Reason    *string `json:"reason"`
}
