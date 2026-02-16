package authMiddlewares

import (
	"i9chat/src/appTypes"
	"i9chat/src/helpers"
	"i9chat/src/services/securityServices"
	"os"

	"github.com/gofiber/fiber/v3"
)

func UserAuth(c fiber.Ctx) error {
	usData := helpers.FromMsgPack[map[string]any](c.Cookies("session"))["user"]

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
