package routes

import (
	"github.com/amanguptak/fiber-api/dtos"
	"github.com/amanguptak/fiber-api/helpers"
	"github.com/amanguptak/fiber-api/models"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
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
