package authMiddlewares

import (
	"encoding/base64"
	"i9chat/src/appTypes"
	"i9chat/src/helpers"
	"i9chat/src/services/securityServices"
	"os"

	"github.com/gofiber/fiber/v3"
)

func UserAuth(c fiber.Ctx) error {
	sess := c.Cookies("session")
	if sess == "" {
		return c.Status(fiber.StatusUnauthorized).SendString("authentication required")
	}

	val, err := base64.RawURLEncoding.DecodeString(sess)
	if err != nil {
		return err
	}

	usData := helpers.FromBtMsgPack[map[string]any](val)["user"]

	if usData == nil {
		return c.Status(fiber.StatusUnauthorized).SendString("authentication required")
	}

	sessionToken := usData.(map[string]any)["authJwt"].(string)

	clientUser, err := securityServices.JwtVerify[appTypes.ClientUser](sessionToken, os.Getenv("AUTH_JWT_SECRET"))
	if err != nil {
		return err
	}

	c.Locals("user", clientUser)

	return c.Next()
}
