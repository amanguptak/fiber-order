package handlers

import (
	"errors"

	"github.com/amanguptak/fiber-api/dtos"
	"github.com/amanguptak/fiber-api/helpers"
	"github.com/amanguptak/fiber-api/models"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func CreateUser(c *fiber.Ctx) error {
	// var user models.User

	var userDto dtos.User

	// BodyParser reads the request body (JSON) and converts it into the 'user' struct.
	// It matches JSON keys (e.g., "firstName") to struct tags.
	// 1. Parse Body into DTO (which has validation tags we change from usermodel type to this because dto has validation tag also  the comment alreay mention in dto file both reason

	if err := c.BodyParser(&userDto); err != nil {
		// If parsing fails (e.g., invalid JSON), return 400 Bad Request and the error message.
		return c.Status(fiber.StatusBadRequest).JSON(err.Error())
	}
	// Why BodyParser REQUIRES &userDto (pointer): BodyParser needs to modify the userDto variable (fill it with data from the request). In Go, if you pass a value without &, the function gets a copy, so any changes it makes won't affect your original variable. By passing &userDto, you give BodyParser the memory address so it can directly write the data into your variable.

	// Why validate.Struct DOESN'T require &userDto: The validator only reads the struct fields to check if they meet the rules. It doesn't modify anything, so it doesn't need a pointer. It can work with either a copy or the original.

	// Summary:

	// BodyParser(&userDto) → Must use & (needs to write data)
	// validate.Struct(userDto) → & is optional (only reads data)

	// validate the DTO

	validate := validator.New()

	if err := validate.Struct(userDto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(helpers.FormatValidationErrors(err))
	}

	user := models.User{
		FirstName: userDto.FirstName,
		LastName:  userDto.LastName,
	}

	// db.Create(&user) inserts a new row into the database.
	// It also runs the BeforeCreate hook to generate the UUID.
	helpers.DB().Create(&user)
	// without createResponseUser we need to write every where like this response := UserResponse{ Id: user.ID.String(), FirstName: user.FirstName, LastName: user.LastName }

	// CreateResponseUser maps the DB model to a safe Response struct (DTO)
	// to hide sensitive fields and ensure consistent API output.
	responseUser := dtos.CreateResponseUser(user)

	return c.Status(fiber.StatusOK).JSON(responseUser)
}

func GetUsers(c *fiber.Ctx) error {
	users := []models.User{}

	// Check for database errors
	if err := helpers.DB().Find(&users).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	responseUsers := make([]dtos.User, 0, len(users))

	// 	This is called Slice Pre-allocation.

	// make([]Type, length, capacity)

	// []dtos.User: The type of the slice.
	// 0 (Length): The slice starts empty (contains 0 items).
	// len(users) (Capacity): We reserve memory for exactly N items (where N is the number of users we found in the DB).
	// Why do we do this? If we found 100 users in the DB, we know we will create exactly 100 response objects. By telling Go "Hey, reserve space for 100 items right now," Go allocates that memory once.

	// If we didn't do this ([]dtos.User{}), Go would have to:

	// Create a tiny array.
	// Fill it up.
	// Realize it's full.
	// Create a bigger array.
	// Copy everything over.
	// Repeat...
	// Pre-allocation avoids all that resizing work, making your code faster and more efficient.

	for _, user := range users {
		responseUser := dtos.CreateResponseUser(user)
		responseUsers = append(responseUsers, responseUser)
	}

	return c.Status(fiber.StatusOK).JSON(responseUsers)
}

//  You are asking why we fetch data into models.User instead of dtos.User in the
// GetUsers
//  function.

// The Reason: GORM (helpers.DB().Find(...)) is an ORM (Object-Relational Mapper). It maps database tables directly to Go structs.

// Database Connection: Your models.User struct has the GORM tags (like gorm:"primaryKey") that tell GORM how to talk to the users table.
// DTO Disconnection: Your dtos.User struct does not have these GORM tags. It only has JSON and validation tags. GORM doesn't know how to map the database columns to your DTO fields.
// So the flow MUST be:

// Database -> Model (using GORM tags)
// Model -> DTO (using your mapper function)
// DTO -> JSON Response
// If you tried helpers.DB().Find(&userDtos), GORM would likely fail or return empty results because it wouldn't know which table or columns to look at.

func findUser(id string, user *models.User) error {
	helpers.DB().Find(user, "id = ?", id)

	if user.ID == uuid.Nil {
		return errors.New("user does not exist")
	}
	return nil
}

func GetUser(c *fiber.Ctx) error {
	id := c.Params("id")

	user := models.User{}

	err := findUser(id, &user)
	if err != nil {
		return c.Status(400).JSON(err.Error())
	}
	responseUser := dtos.CreateResponseUser(user)
	return c.Status(fiber.StatusOK).JSON(responseUser)

}

func UpdateUser(c *fiber.Ctx) error {
	id := c.Params("id")
	updatedUser := dtos.UpdateUser{}
	if err := c.BodyParser(&updatedUser); err != nil {
		// If parsing fails (e.g., invalid JSON), return 400 Bad Request and the error message.
		return c.Status(fiber.StatusBadRequest).JSON(err.Error())
	}

	validate := validator.New()

	if err := validate.Struct(updatedUser); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(helpers.FormatValidationErrors(err))
	}
	user := models.User{}

	err := findUser(id, &user)
	if err != nil {
		return c.Status(400).JSON(err.Error())
	}
	// Only update FirstName if it was actually sent in the JSON
	if updatedUser.FirstName != nil {
		user.FirstName = *updatedUser.FirstName
	}

	// Only update LastName if it was actually sent in the JSON
	if updatedUser.LastName != nil {
		user.LastName = *updatedUser.LastName
	}
	helpers.DB().Save(&user)
	responseUser := dtos.CreateResponseUser(user)
	return c.Status(fiber.StatusOK).JSON(responseUser)

}

func DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	user := models.User{}

	err := findUser(id, &user)
	if err != nil {
		return c.Status(400).JSON(err.Error())
	}

	// This is a compact Go syntax called the "If with Short Statement".
	if err = helpers.DB().Delete(&user).Error; err != nil {
		return c.Status(400).JSON(err.Error())
	}

	// 	Execute: err = helpers.DB().Delete(&user).Error (Run the delete command and assign the result to err)
	// Check: err != nil (Check if that error is not nil)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "User deleted successfully"})

}
