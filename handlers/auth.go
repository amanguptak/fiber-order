package handlers

import (
	"time"

	"github.com/amanguptak/fiber-api/dtos"
	"github.com/amanguptak/fiber-api/helpers"
	"github.com/amanguptak/fiber-api/models"
	"github.com/amanguptak/fiber-api/repositories"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func Register(c *fiber.Ctx) error {
	var data dtos.RegisterRequest

	if err := c.BodyParser(&data); err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	validate := validator.New()

	if err := validate.Struct(data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(helpers.FormatValidationErrors(err))
	}

	password, _ := bcrypt.GenerateFromPassword([]byte(data.Password), 12)

	user := models.User{
		FirstName: data.FirstName,
		LastName:  data.LastName,
		Email:     data.Email,
		Password:  password,
	}

	if err := helpers.DB().Create(&user).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			// "error":err.Error(),
			"error": "Could not create user"})
	}
	responseUser := dtos.CreateResponseUser(user)
	return c.Status(fiber.StatusOK).JSON(responseUser)
}

func Login(c *fiber.Ctx) error {
	var data dtos.LoginRequest

	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid data",
		})
	}

	validate := validator.New()
	if err := validate.Struct(data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(helpers.FormatValidationErrors(err))
	}

	var user models.User

	helpers.DB().Where("email = ?", data.Email).First(&user)

	if user.ID == uuid.Nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(data.Password)); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid credentials",
		})
	}
	// USE HELPER: Generate Access Token (15 mins)
	// We use our helper to create a short-lived token for API access.

	token, _ := helpers.GenerateToken(user.ID.String(), time.Now().Add(time.Minute*15))

	// USE HELPER: Generate Refresh Token (7 days)
	// We create a long-lived token so the user doesn't have to login every 15 mins.

	refreshToken, _ := helpers.GenerateToken(user.ID.String(), time.Now().Add(time.Hour*24*7))

	if err := repositories.StoreRefreshToken(user.ID, refreshToken, time.Now().Add(7*24*time.Hour)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Set Cookie
	// We put the Refresh Token in an HttpOnly cookie
	cookie := fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Expires:  time.Now().Add(time.Hour * 24 * 7),
		HTTPOnly: true,  // CRITICAL: JavaScript cannot read this. Prevents XSS attacks.
		SameSite: "Lax", // CSRF protection
		Secure:   false, // Set to true in production (HTTPS only)
	}
	c.Cookie(&cookie)
	return c.JSON(fiber.Map{
		"message": "Login successfully",
		"token":   token,
	})
}

func Logout(c *fiber.Ctx) error {
	cookie := c.Cookies("refresh_token")

	// Mark token as revoked in DB (best practice)
	hash := repositories.HashToken(cookie)
	helpers.DB().Model(&models.RefreshToken{}).
		Where("token_hash = ?", hash).
		Update("is_revoked", true)

	// Clear cookie
	c.ClearCookie("refresh_token")
	return c.JSON(fiber.Map{"message": "Logged out successfully"})
}

func Refresh(c *fiber.Ctx) error {
	cookie := c.Cookies("refresh_token")
	token, err := helpers.ParseToken(cookie)
	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "unauthenticated",
		})
	}

	// claims := token.Claims.(*jwt.MapClaims)

	// 	We had to extract the User ID from the JWT claims because we didn't have a database to look it up.

	// But now with Rotation: The database IS the source of truth. We don't trust the JWT claims alone anymore.

	// ✅ Get BOTH tokens
	newAccessToken, newRefreshToken, err := repositories.RotateRefreshToken(cookie)
	if err != nil {
		c.ClearCookie("refresh_token")
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Invalid token"})
	}
	// ✅ Send NEW Refresh Token as HttpOnly cookie
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    newRefreshToken,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HTTPOnly: true,
		SameSite: "Lax", // CSRF protection
		Secure:   false, // Set to true in production (HTTPS only)
	})

	// ✅ Return NEW Access Token in JSON
	return c.JSON(fiber.Map{
		"token": newAccessToken,
	})
}
