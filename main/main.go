package main

import (
	authroutes "routes/auth"
	userroutes "routes/user"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	app.Route("/api/auth", authroutes.Init)
	app.Route("/api/user", userroutes.Init)
}
