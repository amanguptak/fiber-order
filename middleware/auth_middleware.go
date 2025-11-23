package middleware

import (
	"github.com/amanguptak/fiber-api/helpers"
	"github.com/gofiber/fiber/v2"
)

func IsAuthenticated(c *fiber.Ctx) error {

	authHeader := c.Get("Authorization") //  Get from Authorization header
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "unauthenticated"})
	}
	tokenString := authHeader[7:]
	token, err := helpers.ParseToken(tokenString)

	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "unauthenticated",
		})
	}

	return c.Next()
}
