// Signup Session Middleware
package ssm

import (
	"encoding/json"
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

	ssbt := sess.Get("signup_session").([]byte)

	var signupSession appTypes.SignupSession

	if err := json.Unmarshal(ssbt, &signupSession); err != nil {
		log.Println("ssm.go: VerifyEmail: json.Unmarshal:", err)
		return fiber.ErrInternalServerError
	}

	if signupSession.Step != "verify email" {
		return c.Status(fiber.StatusUnauthorized).SendString("session error")
	}

	c.Locals("signup_session_data", signupSession.Data)

	return c.Next()
}

func RegisterUser(c *fiber.Ctx) error {
	sess, err := appGlobals.SignupSessionStore.Get(c)
	if err != nil {
		log.Println("ssm.go: RegisterUser: SignupSessionStore.Get:", err)
		return fiber.ErrInternalServerError
	}

	ssbt := sess.Get("signup_session").([]byte)

	var signupSession appTypes.SignupSession

	if err := json.Unmarshal(ssbt, &signupSession); err != nil {
		log.Println("ssm.go: RegisterUser: json.Unmarshal:", err)
		return fiber.ErrInternalServerError
	}

	if signupSession.Step != "register user" {
		return c.Status(fiber.StatusUnauthorized).SendString("session error")
	}

	c.Locals("signup_session_data", signupSession.Data)

	return c.Next()
}
