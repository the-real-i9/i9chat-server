package middlewares

import (
	"encoding/json"
	"i9chat/appGlobals"
	"log"

	"github.com/gofiber/fiber/v2"
)

func ValidateSession(c *fiber.Ctx) error {
	sess, err := appGlobals.SessionStore.Get(c)
	if err != nil {
		log.Println("ssm.go: ValidateSession: SessionStore.Get:", err)
		return fiber.ErrInternalServerError
	}

	ssbt, ok := sess.Get("signup").([]byte)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).SendString("wrong endpoint access: complete the previous step of the signup process")
	}

	var signupSessionData map[string]any

	if err := json.Unmarshal(ssbt, &signupSessionData); err != nil {
		log.Println("ssm.go: ValidateSession: json.Unmarshal:", err)
		return fiber.ErrInternalServerError
	}

	c.Locals("signup_sess_data", signupSessionData)

	return c.Next()
}
