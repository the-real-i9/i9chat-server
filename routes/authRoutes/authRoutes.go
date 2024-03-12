package authroutes

import (
	authcontrollers "controllers/auth"

	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	router.Get("/signup/request_new_account", authcontrollers.RequestNewAccount)
}
