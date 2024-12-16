package middlewares

import (
	"i9chat/appTypes"
	"i9chat/services/securityServices"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func SignupSession(c *fiber.Ctx) error {

	sessionToken := c.Get("Authorization")

	sessionData, err := securityServices.JwtVerify[appTypes.SignupSessionData](sessionToken, os.Getenv("SIGNUP_SESSION_JWT_SECRET"))
	if err != nil {
		return err
	}

	c.Locals("session", sessionData)

	return c.Next()
}

func Auth(c *fiber.Ctx) error {
	authHeaderValue := c.Get("Authorization")

	sessionToken := strings.Fields(authHeaderValue)[1]

	clientUser, err := securityServices.JwtVerify[appTypes.ClientUser](sessionToken, os.Getenv("AUTH_JWT_SECRET"))
	if err != nil {
		return err
	}

	c.Locals("user", clientUser)

	return c.Next()
}
