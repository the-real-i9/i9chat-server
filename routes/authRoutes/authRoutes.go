package authRoutes

import (
	"i9chat/controllers/authControllers"
	"i9chat/middlewares"
	"os"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	router.Get("/signup/request_new_account", authControllers.RequestNewAccount)

	router.Get("/signin", authControllers.Signin)

	router.Use(jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(os.Getenv("SIGNUP_SESSION_JWT_SECRET"))},
		ContextKey: "auth",
	}))

	router.Get("/signup/verify_email", middlewares.VerifyEmail, authControllers.VerifyEmail)

	router.Get("/signup/register_user", middlewares.RegisterUser, authControllers.RegisterUser)
}
