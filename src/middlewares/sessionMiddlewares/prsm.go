package sessionMiddlewares

import (
	"i9chat/src/helpers"

	"github.com/gofiber/fiber/v3"
	"github.com/vmihailenco/msgpack/v5"
)

func PasswordResetSession(c fiber.Ctx) error {
	prsData := helpers.FromMsgPack[map[string]msgpack.RawMessage](c.Cookies("session"))["passwordReset"]

	if prsData == nil {
		return c.Status(fiber.StatusUnauthorized).SendString("no ongoing password reset session or this session endpoint was accessed out of order")
	}

	c.Locals("passwordReset_sess_data", prsData)

	return c.Next()
}
