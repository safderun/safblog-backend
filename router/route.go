package router

import (
	"fmt"
	"safblog-backend/controllers"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	// grouping
	fmt.Println("Setting up routes")
	fmt.Println("Getting routes ready.")

	api := app.Group("/api")
	version := api.Group("/v1")

	// routes
	auth := version.Group("/auth")
	auth.Post("/register", controllers.RegisterController)
	auth.Post("/login", controllers.LoginController)
}
