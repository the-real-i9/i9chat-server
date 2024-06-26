package main

import (
	"i9chat/routes/authRoutes"
	"i9chat/routes/chatRoutes"
	"i9chat/routes/userRoutes"
	"i9chat/utils/helpers"
	"log"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

func main() {
	if err := helpers.LoadEnv(); err != nil {
		log.Fatal(err)
	}

	if err := helpers.InitDBPool(); err != nil {
		log.Fatal(err)
	}

	if err := helpers.InitGCSClient(); err != nil {
		log.Fatal(err)
	}

	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	app.Use("/", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}

		return fiber.ErrUpgradeRequired
	})

	app.Route("/api/auth", authRoutes.Init)

	app.Route("/api/app/user", userRoutes.Init)

	app.Route("/api/app/dm_chat", chatRoutes.InitDMChat)
	app.Route("/api/app/group_chat", chatRoutes.InitGroupChat)

	log.Fatalln(app.Listen("localhost:8000"))
}
