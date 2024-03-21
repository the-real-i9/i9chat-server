package main

import (
	"log"
	"routes/authroutes"
	"routes/chatroutes"
	"routes/userroutes"
	"utils/appglobals"
	"utils/helpers"

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

	if err := appglobals.InitGCSClient(); err != nil {
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

	app.Route("/api/app/user", userroutes.Init)

	app.Route("/api/app/dm_chat", chatroutes.InitDMChat)
	app.Route("/api/app/group_chat", chatroutes.InitGroupChat)

	log.Fatalln(app.Listen("localhost:8000"))
}
