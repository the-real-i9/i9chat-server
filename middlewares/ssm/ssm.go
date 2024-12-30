// Signup Session Middleware
package ssm

import (
	"i9chat/appGlobals"
	"i9chat/appTypes"
	"log"

	"github.com/gofiber/fiber/v2"
)

func VerifyEmail(c *fiber.Ctx) error {
	sess, err := appGlobals.SignupSessionStore.Get(c)
	if err != nil {
		log.Println("ssm.go: VerifyEmail: SignupSessionStore.Get:", err)
		return fiber.ErrInternalServerError
	}

	signupSession := sess.Get("signup_session").(*appTypes.SignupSession)

	if signupSession.Step != "verify email" {
		return c.Status(fiber.StatusUnauthorized).SendString("session error")
	}

	c.Locals("session", sess)

	return c.Next()
}

func RegisterUser(c *fiber.Ctx) error {
	sess, err := appGlobals.SignupSessionStore.Get(c)
	if err != nil {
		log.Println("ssm.go: RegisterUser: SignupSessionStore.Get:", err)
		return fiber.ErrInternalServerError
	}

	signupSession := sess.Get("session").(*appTypes.SignupSession)

	if signupSession.Step != "register user" {
		return c.Status(fiber.StatusUnauthorized).SendString("session error")
	}

	c.Locals("session", sess)

	return c.Next()
}
