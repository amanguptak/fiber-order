package main

import (
	"log"

	"github.com/amanguptak/fiber-api/database"

	"github.com/amanguptak/fiber-api/routes"
	"github.com/gofiber/fiber/v2"
)

func main() {
	database.ConnectDb()
	app := fiber.New()
	routes.SetupRoutes(app)

	log.Fatal(app.Listen(":8000"))
}
