package main

import (
	"i9chat/src/initializers"
	"i9chat/src/routes/appRoutes"
	"i9chat/src/routes/authRoutes"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/encryptcookie"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func init() {
	if err := initializers.InitApp(); err != nil {
		log.Fatal(err)
	}
}

//	@title			i9chat Chat App API
//	@version		1.0
//	@description	i9chat Chat App Backend API.

//	@contact.name	i9ine
//	@contact.email	oluwarinolasam@gmail.com

//	@host		localhost:8000
//	@BasePath	/api

//	@securityDefinitions.apikey	ApiKeyAuth
//	@in							header
//	@name						Cookie
//	@description				JWT API key in encrypted cookie to protect private endpoints

//	@accepts	json
//	@produces	json

// @schemes	http https
func main() {
	defer initializers.CleanUp()

	app := fiber.New()

	app.Use(limiter.New())

	app.Use(helmet.New(helmet.Config{
		// CrossOriginResourcePolicy: "cross-origin", /* for production */
	}))

	app.Use(cors.New(cors.Config{
		// AllowOrigins:     "http://localhost:5173", /* production client host */
		// AllowCredentials: true,
	}))

	app.Use(encryptcookie.New(encryptcookie.Config{
		Key: os.Getenv("COOKIE_SECRET"),
	}))

	app.Route("/api/auth", authRoutes.Route)

	app.Route("/api/app", appRoutes.Route)

	var PORT string

	if os.Getenv("GO_ENV") != "production" {
		PORT = "8000"
	} else {
		PORT = os.Getenv("PORT")
	}

	log.Fatalln(app.Listen("0.0.0.0:" + PORT))
}
