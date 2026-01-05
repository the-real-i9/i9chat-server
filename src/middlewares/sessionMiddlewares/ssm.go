package sessionMiddlewares

import (
	"i9chat/src/helpers"

	"github.com/gofiber/fiber/v2"
)

func SignupSession(c *fiber.Ctx) error {
	ssData := helpers.FromJson[map[string]any](c.Cookies("session"))["signup"]

	if ssData == nil {
		return c.Status(fiber.StatusUnauthorized).SendString("no ongoing signup session or this session endpoint was accessed out of order")
	}

	c.Locals("signup_sess_data", ssData)

	return c.Next()
}
