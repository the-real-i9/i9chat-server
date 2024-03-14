package authroutes

import (
	"controllers/authcontrollers"

	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	router.Get("/signup/request_new_account", authcontrollers.RequestNewAccount)

	router.Get("/signup/verify_email", authcontrollers.VerifyEmail)

	router.Get("/signup/register_user", authcontrollers.RegisterUser)

	router.Get("/signin", authcontrollers.Signin)
}
