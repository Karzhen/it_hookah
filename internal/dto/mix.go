package dto

type MixItemInput struct {
	ProductID string `json:"product_id"`
	Percent   int16  `json:"percent"`
}

type CreateMixRequest struct {
	Name               string         `json:"name"`
	Description        *string        `json:"description"`
	FinalStrengthLabel *string        `json:"final_strength_label"`
	IsActive           *bool          `json:"is_active"`
	Items              []MixItemInput `json:"items"`
}

type UpdateMixRequest struct {
	Name               *string         `json:"name"`
	Description        *string         `json:"description"`
	FinalStrengthLabel *string         `json:"final_strength_label"`
	IsActive           *bool           `json:"is_active"`
	Items              *[]MixItemInput `json:"items"`
}

type MixItemResponse struct {
	ProductID   string `json:"product_id"`
	ProductName string `json:"product_name"`
	Percent     int16  `json:"percent"`
}

type MixTagResponse struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type MixResponse struct {
	ID                 string            `json:"id"`
	Name               string            `json:"name"`
	Description        *string           `json:"description,omitempty"`
	FinalStrengthLabel *string           `json:"final_strength_label,omitempty"`
	IsActive           bool              `json:"is_active"`
	Items              []MixItemResponse `json:"items"`
	Tags               []MixTagResponse  `json:"tags"`
	CreatedAt          string            `json:"created_at"`
	UpdatedAt          string            `json:"updated_at"`
}
