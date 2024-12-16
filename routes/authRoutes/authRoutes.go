package authRoutes

import (
	"i9chat/controllers/authControllers/signinControllers"
	"i9chat/controllers/authControllers/signupControllers"
	"i9chat/middlewares"

	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	ssm := middlewares.SignupSession

	router.Get("/signup/request_new_account", signupControllers.RequestNewAccount)

	router.Get("/signup/verify_email", ssm, signupControllers.VerifyEmail)

	router.Get("/signup/register_user", ssm, signupControllers.RegisterUser)

	router.Get("/signin", signinControllers.Signin)
}
