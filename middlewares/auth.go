package middlewares

import (
	"encoding/json"
	"i9chat/appGlobals"
	"i9chat/appTypes"
	"i9chat/services/securityServices"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
)

func UserAuth(c *fiber.Ctx) error {
	sess, err := appGlobals.SessionStore.Get(c)
	if err != nil {
		log.Println("auth.go: UserAuth: SessionStore.Get:", err)
		return fiber.ErrInternalServerError
	}

	usbt, ok := sess.Get("user").([]byte)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).SendString("authentication required")
	}

	var userSessionData map[string]string

	if err := json.Unmarshal(usbt, &userSessionData); err != nil {
		log.Println("auth.go: UserAuth: json.Unmarshal:", err)
		return fiber.ErrInternalServerError
	}

	sessionToken := userSessionData["authJwt"]

	clientUser, err := securityServices.JwtVerify[appTypes.ClientUser](sessionToken, os.Getenv("AUTH_JWT_SECRET"))
	if err != nil {
		return err
	}

	c.Locals("user", clientUser)

	return c.Next()
}
