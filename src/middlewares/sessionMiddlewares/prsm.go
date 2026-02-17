package sessionMiddlewares

import (
	"encoding/base64"
	"i9chat/src/helpers"

	"github.com/gofiber/fiber/v3"
	"github.com/vmihailenco/msgpack/v5"
)

func PasswordResetSession(c fiber.Ctx) error {
	sess := c.Cookies("session")
	if sess == "" {
		return c.Status(fiber.StatusUnauthorized).SendString("no ongoing password reset session")
	}

	val, err := base64.RawURLEncoding.DecodeString(sess)
	if err != nil {
		return err
	}

	prsData := helpers.FromBtMsgPack[map[string]msgpack.RawMessage](val)["passwordReset"]

	if prsData == nil {
		return c.Status(fiber.StatusUnauthorized).SendString("no ongoing password reset session or this session endpoint was accessed out of order")
	}

	c.Locals("passwordReset_sess_data", prsData)

	return c.Next()
}
