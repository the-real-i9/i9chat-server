package sessionMiddlewares

import (
	"i9chat/src/helpers"

	"github.com/gofiber/fiber/v2"
)

func SignupSession(c *fiber.Ctx) error {
	ssStr := c.Cookies("signup")

	if ssStr == "" {
		return c.Status(fiber.StatusUnauthorized).SendString("out-of-turn endpoint access: complete the previous step of the signup process")
	}

	signupSessionData := helpers.FromJson[map[string]any](ssStr)

	c.Locals("signup_sess_data", signupSessionData)

	return c.Next()
}
