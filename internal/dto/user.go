package dto

type UserProfileResponse struct {
	ID         string  `json:"id"`
	Email      string  `json:"email"`
	FirstName  string  `json:"first_name"`
	LastName   string  `json:"last_name"`
	MiddleName *string `json:"middle_name,omitempty"`
	Phone      *string `json:"phone,omitempty"`
	Age        *int    `json:"age,omitempty"`
	Role       string  `json:"role"`
	IsActive   bool    `json:"is_active"`
	CreatedAt  string  `json:"created_at"`
}

type UpdateMeRequest struct {
	FirstName  *string `json:"first_name"`
	LastName   *string `json:"last_name"`
	MiddleName *string `json:"middle_name"`
	Phone      *string `json:"phone"`
	Age        *int    `json:"age"`
}
