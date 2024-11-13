package authRoutes

import (
	"i9chat/controllers/authControllers/signinControllers"
	"i9chat/controllers/authControllers/signupControllers"

	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	router.Get("/signup/request_new_account", signupControllers.RequestNewAccount)

	router.Get("/signup/verify_email", signupControllers.VerifyEmail)

	router.Get("/signup/register_user", signupControllers.RegisterUser)

	router.Get("/signin", signinControllers.Signin)
}
