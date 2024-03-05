package main

import (
	"log"
	authroutes "routes/auth"
	"utils/helpers"

	"github.com/gofiber/fiber/v2"
)

func main() {
	if err := helpers.LoadEnv(); err != nil {
		log.Fatal(err)
	}

	app := fiber.New()

	app.Route("/api/auth", authroutes.Init)
	// app.Route("/api/user", userroutes.Init)

	log.Fatalln(app.Listen("localhost:8080"))
}
