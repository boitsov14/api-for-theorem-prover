package main

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	_ "github.com/joho/godotenv/autoload"
)

func main() {

	app := fiber.New()

	app.Use(cors.New())

	app.Post("/twitter", twitterBot)
	app.Post("/web", webApp)

	// portの設定
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	app.Listen(":" + port)
}
