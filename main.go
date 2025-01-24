package main

import (
	"i9chat/initializers"
	"i9chat/middlewares"
	"i9chat/routes/authRoutes"
	"i9chat/routes/chatRoutes/dmChatRoutes"
	"i9chat/routes/chatRoutes/groupChatRoutes"
	"i9chat/routes/userRoutes"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/helmet"
)

func init() {
	if err := initializers.InitApp(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	defer func() {
		initializers.CleanUp()
	}()

	app := fiber.New()

	app.Use(helmet.New())

	app.Route("/api/auth", authRoutes.Init)

	app.Use("/api/app", middlewares.Auth)

	app.Route("/api/app/user", userRoutes.Init)

	app.Route("/api/app/dm_chat", dmChatRoutes.Init)
	app.Route("/api/app/group_chat", groupChatRoutes.Init)

	var PORT string

	if os.Getenv("GO_ENV") != "production" {
		PORT = "8000"
	} else {
		PORT = os.Getenv("PORT")
	}

	log.Fatalln(app.Listen("0.0.0.0:" + PORT))

}
