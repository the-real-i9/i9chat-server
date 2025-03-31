package authRoutes

import (
	"i9chat/src/controllers/authControllers/signinControllers"
	"i9chat/src/controllers/authControllers/signupControllers"
	ssm "i9chat/src/middlewares"

	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	router.Post("/signup/request_new_account", signupControllers.RequestNewAccount)

	router.Post("/signup/verify_email", ssm.ValidateSession, signupControllers.VerifyEmail)

	router.Post("/signup/register_user", ssm.ValidateSession, signupControllers.RegisterUser)

	router.Post("/signin", signinControllers.Signin)
}
