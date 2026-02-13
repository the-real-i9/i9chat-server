package sessionMiddlewares

import (
	"i9chat/src/helpers"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
)

func SignupSession(c *fiber.Ctx) error {
	ssData := helpers.FromJson[map[string]json.RawMessage](c.Cookies("session"))["signup"]

	if ssData == nil {
		return c.Status(fiber.StatusUnauthorized).SendString("no ongoing signup session or this session endpoint was accessed out of order")
	}

	c.Locals("signup_sess_data", ssData)

	return c.Next()
}
