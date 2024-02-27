package main

import (
	"log"
	authroutes "routes/auth"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	app.Route("/api/auth", authroutes.Init)
	// app.Route("/api/user", userroutes.Init)

	log.Fatalln(app.Listen("localhost:8000"))
}
