package main

import (
	"i9chat/src/initializers"
	"i9chat/src/routes/appRoutes"
	"i9chat/src/routes/authRoutes"
	"log"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/encryptcookie"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/vmihailenco/msgpack/v5"
)

func init() {
	if err := initializers.InitApp(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	defer initializers.CleanUp()

	app := fiber.New(fiber.Config{
		MsgPackEncoder: msgpack.Marshal,
		MsgPackDecoder: msgpack.Unmarshal,
	})

	app.Use(limiter.New())

	app.Use(helmet.New(helmet.Config{
		// CrossOriginResourcePolicy: "cross-origin", /* for production */
	}))

	app.Use(cors.New(cors.Config{
		// AllowOrigins:     []string{"http://localhost:5173"}, /* production client host */
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
