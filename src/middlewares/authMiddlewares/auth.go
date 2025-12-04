package authMiddlewares

import (
	"i9chat/src/appTypes"
	"i9chat/src/helpers"
	"i9chat/src/services/securityServices"
	"os"

	"github.com/gofiber/fiber/v2"
)

func UserAuth(c *fiber.Ctx) error {
	usStr := c.Cookies("user")

	if usStr == "" {
		return c.Status(fiber.StatusUnauthorized).SendString("authentication required")
	}

	userSessionData := helpers.FromJson[map[string]string](usStr)

	sessionToken := userSessionData["authJwt"]

	clientUser, err := securityServices.JwtVerify[appTypes.ClientUser](sessionToken, os.Getenv("AUTH_JWT_SECRET"))
	if err != nil {
		return err
	}

	c.Locals("user", clientUser)

	return c.Next()
}
