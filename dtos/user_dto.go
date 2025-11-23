package dtos

import "github.com/amanguptak/fiber-api/models"

// User struct is a Data Transfer Object (DTO).
// 1. Security: It filters out sensitive DB fields (like passwords) so they are never sent to the user.
// 2. Decoupling: It allows the DB schema to change (e.g., renaming fields) without breaking the API response.
type User struct {
	Id        string `json:"id"`
	FirstName string `json:"firstName" validate:"required,min=2,max=32"`
	LastName  string `json:"lastName" validate:"required ,min=2,max=31"`
	Email     string `json:"email" validate:"required,email"`
}

type UpdateUser struct {
	FirstName *string `json:"firstName" validate:"omitempty,min=2,max=32"`
	LastName  *string `json:"lastName" validate:"omitempty,min=2,max=31"`
	Email     *string `json:"email" validate:"omitempty,email"`
}

// CreateResponseUser is a "Mapper" function.
// It translates the internal Database Model (models.User) into the public Response DTO (routes.User).
// This ensures all endpoints return data in a consistent format.

func CreateResponseUser(user models.User) User {
	return User{
		Id:        user.ID.String(),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
	}
}
