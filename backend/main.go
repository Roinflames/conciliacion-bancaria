package main

import (
	"conciliacion-bancaria/config"
	"conciliacion-bancaria/db"
	"conciliacion-bancaria/routes"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	config.Load()
	db.Init(config.DBPath)

	app := fiber.New(fiber.Config{
		AppName:      "Conciliación Bancaria v0.0.1",
		BodyLimit:    10 * 1024 * 1024, // 10 MB (para archivos Excel grandes)
	})

	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"perfil":  config.Profile,
			"version": "0.0.1",
		})
	})

	// Rutas API
	routes.Register(app)

	// Frontend estático
	app.Static("/", "../frontend")

	// Fallback SPA
	app.Get("/*", func(c *fiber.Ctx) error {
		return c.SendFile("../frontend/index.html")
	})

	log.Printf("Conciliación Bancaria corriendo en http://localhost:%s (perfil: %s)", config.Port, config.Profile)
	log.Fatal(app.Listen(":" + config.Port))
}
