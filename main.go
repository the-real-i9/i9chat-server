package main

import (
	"i9chat/initializers"
	"i9chat/routes/authRoutes"
	"i9chat/routes/chatRoutes/dmChatRoutes"
	"i9chat/routes/chatRoutes/groupChatRoutes"
	"i9chat/routes/userRoutes"
	"log"
	"os"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

func init() {
	if err := initializers.InitApp(); err != nil {
		log.Fatal(err)
	}
}

func main() {

	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}

		return fiber.ErrUpgradeRequired
	})

	app.Route("/api/auth", authRoutes.Init)

	app.Route("/api/app/user", userRoutes.Init)

	app.Route("/api/app/dm_chat", dmChatRoutes.Init)
	app.Route("/api/app/group_chat", groupChatRoutes.Init)

	var PORT string

	if os.Getenv("GO_ENV") != "production" {
		PORT = "5000"
	} else {
		PORT = os.Getenv("PORT")
	}

	log.Fatalln(app.Listen("0.0.0.0:" + PORT))

}
