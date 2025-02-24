package middlewares

import (
	"i9chat/appGlobals"
	"i9chat/appTypes"
	"i9chat/services/securityServices"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
)

func Auth(c *fiber.Ctx) error {
	sess, err := appGlobals.UserSessionStore.Get(c)
	if err != nil {
		log.Println("auth.go: Auth: UserSignupSession.Get:", err)
		return fiber.ErrInternalServerError
	}

	sessionToken, ok := sess.Get("authJwt").(string)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "authentication required")

	}

	clientUser, err := securityServices.JwtVerify[appTypes.ClientUser](sessionToken, os.Getenv("AUTH_JWT_SECRET"))
	if err != nil {
		return err
	}

	c.Locals("user", clientUser)

	return c.Next()
}
