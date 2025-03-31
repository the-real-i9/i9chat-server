package main

import (
	"i9chat/src/initializers"
	"i9chat/src/middlewares"
	"i9chat/src/routes/authRoutes"
	"i9chat/src/routes/chatRoutes/groupChatRoutes"
	"i9chat/src/routes/userRoutes"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/encryptcookie"
	"github.com/gofiber/fiber/v2/middleware/helmet"
)

func init() {
	if err := initializers.InitApp(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	defer initializers.CleanUp()

	app := fiber.New()

	app.Use(helmet.New())
	app.Use(cors.New())

	app.Use(encryptcookie.New(encryptcookie.Config{
		Key: os.Getenv("COOKIE_SECRET"),
	}))

	app.Route("/api/auth", authRoutes.Init)

	app.Use("/api/app", middlewares.UserAuth)

	app.Route("/api/app/user", userRoutes.Init)

	app.Route("/api/app/group_chat", groupChatRoutes.Init)

	var PORT string

	if os.Getenv("GO_ENV") != "production" {
		PORT = "8000"
	} else {
		PORT = os.Getenv("PORT")
	}

	log.Fatalln(app.Listen("0.0.0.0:" + PORT))

}
