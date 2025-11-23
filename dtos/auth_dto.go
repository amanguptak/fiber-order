package dtos

type RegisterRequest struct {
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName" validate:"required"`
	Email     string `json:"email" validate:"required,email"`    // Validate it's a real email format
	Password  string `json:"password" validate:"required,min=6"` // Enforce min length for security
}
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}
