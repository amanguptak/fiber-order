package main

import (
	"log"

	"github.com/amanguptak/fiber-api/database"
	"github.com/amanguptak/fiber-api/routes"
	"github.com/gofiber/fiber/v2"
)

func apiRoutes(app *fiber.App) {

	// return c.JSON(fiber.Map{
	// 	"message": "first end point",
	// })
	app.Post("/api/user", routes.CreateUser)

}

func main() {
	database.ConnectDb()
	app := fiber.New()
	apiRoutes(app)

	log.Fatal(app.Listen(":8000"))
}
