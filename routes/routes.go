package routes

import (
	"main/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func Routes(app *fiber.App) {
	// Middleware
	api := app.Group("/api", logger.New())

	// Auth
	auth := api.Group("/auth")
	auth.Post("/login", handlers.Login)
	auth.Post("/logout", handlers.Logout)

	// Productos
	product := api.Group("/products")
	product.Get("/", handlers.GetAllProducts)
	product.Post("/", handlers.NewProduct)
	product.Put("/:id", handlers.EditProduct)
	product.Delete("/:id", handlers.DeleteProduct)

	// Files
	files := api.Group("/files")
	files.Static("/imgs", "./imgs")
	files.Post("/", handlers.UploadMultiFiles)

	// Users
	user := api.Group("/users")
	user.Get("/", handlers.GetUsers)
	user.Get("/:id", handlers.GetUser)
	user.Post("/", handlers.CreateUser)
	user.Patch("/:id", handlers.UpdateUser)
	user.Patch("/profile/:id", handlers.UpdateProfile)

	// Customers
	customer := api.Group("/customers")
	customer.Get("/", handlers.GetCustomers)
	customer.Get("/:id", handlers.GetCustomer)
	customer.Post("/", handlers.CreateCustomer)
	customer.Patch("/:id", handlers.UpdateCustomer)
	customer.Delete("/:id", handlers.DeleteCustomer)
}
