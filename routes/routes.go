package routes



import (
    "github.com/amanguptak/fiber-api/handlers"
    "github.com/amanguptak/fiber-api/middleware"
    "github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
    // Public routes (no authentication required)
    app.Post("/api/register", handlers.Register)
    app.Post("/api/login", handlers.Login)
    app.Post("/api/logout", handlers.Logout)
    app.Post("/api/refresh", handlers.Refresh)

    // Protected routes (authentication required)
    api := app.Group("/api", middleware.IsAuthenticated)
    api.Get("/users/:id", handlers.GetUser)
    api.Patch("/users/:id", handlers.UpdateUser)
    api.Delete("/users/:id", handlers.DeleteUser)
}