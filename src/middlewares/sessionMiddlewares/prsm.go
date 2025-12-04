package sessionMiddlewares

import (
	"i9chat/src/helpers"

	"github.com/gofiber/fiber/v2"
)

func PasswordResetSession(c *fiber.Ctx) error {
	prsStr := c.Cookies("passwordReset")

	if prsStr == "" {
		return c.Status(fiber.StatusUnauthorized).SendString("out-of-turn endpoint access: complete the previous step of the password reset process")
	}

	passwordResetSessionData := helpers.FromJson[map[string]any](prsStr)

	c.Locals("passwordReset_sess_data", passwordResetSessionData)

	return c.Next()
}
