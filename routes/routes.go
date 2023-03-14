package routes

import (
	"main/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func Routes(app *fiber.App) {
	// Middleware
	api := app.Group("/api", logger.New())

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
}
