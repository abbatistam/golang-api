package main

import (
	"log"
	"main/database"
	"main/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	// Conectar a la BD
	if err := database.Connect(); err != nil {
		log.Fatal(err)
	}

	// Fiber app
	app := fiber.New()
	app.Use(cors.New())

	// Rutas
	routes.Routes(app)

	log.Fatal(app.Listen(":3000"))
}
