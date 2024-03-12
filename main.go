package main

import (
	"log"
	authroutes "routes/auth"
	"utils/helpers"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

func main() {
	if err := helpers.LoadEnv(); err != nil {
		log.Fatal(err)
	}

	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	app.Use("/", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}

		return fiber.ErrUpgradeRequired
	})

	app.Route("/api/auth", authroutes.Init)

	/* jwtToken := helpers.GenerateJwtToken(map[string]any{"email": "oluwarinolasam@gmail.com"})

	fmt.Println(jwtToken) */

	log.Fatalln(app.Listen("localhost:8080"))
}
