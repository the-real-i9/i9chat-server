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

	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	app.Route("/api/auth", authroutes.Init)
	// app.Route("/api/user", userroutes.Init)
	/* user := usermodel.User{Id: 2}
	data, err := user.Edit([][]string{{"prece", "fhunmytor17"}})
	if err != nil {
		log.Println(err)
	}
	json_data, _ := json.MarshalIndent(data, "", " ")
	os.Stdout.Write(json_data) */
	/* exist, err := usermodel.UserExists("i9")
	if err != nil {
		log.Println(err)
	}
	log.Println(exist) */

	log.Fatalln(app.Listen("localhost:8080"))
}
