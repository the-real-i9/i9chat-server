package sessionMiddlewares

import (
	"encoding/base64"
	"i9chat/src/helpers"

	"github.com/gofiber/fiber/v3"
	"github.com/vmihailenco/msgpack/v5"
)

func SignupSession(c fiber.Ctx) error {
	sess := c.Cookies("session")
	if sess == "" {
		return c.Status(fiber.StatusUnauthorized).SendString("no ongoing password reset session")
	}

	val, err := base64.RawURLEncoding.DecodeString(sess)
	if err != nil {
		return err
	}

	ssData := helpers.FromBtMsgPack[map[string]msgpack.RawMessage](val)["signup"]

	if ssData == nil {
		return c.Status(fiber.StatusUnauthorized).SendString("no ongoing signup session or this session endpoint was accessed out of order")
	}

	c.Locals("signup_sess_data", ssData)

	return c.Next()
}
