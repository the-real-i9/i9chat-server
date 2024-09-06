package authRoutes

import (
	"i9chat/src/controllers/authControllers"

	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	router.Get("/signup/request_new_account", authControllers.RequestNewAccount)

	router.Get("/signup/verify_email", authControllers.VerifyEmail)

	router.Get("/signup/register_user", authControllers.RegisterUser)

	router.Get("/signin", authControllers.Signin)
}
