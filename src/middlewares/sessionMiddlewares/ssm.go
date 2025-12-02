package sessionMiddlewares

import (
	"log"

	"github.com/goccy/go-json"

	"github.com/gofiber/fiber/v2"
)

func SignupSession(c *fiber.Ctx) error {
	ssStr := c.Cookies("signup")

	if ssStr == "" {
		return c.Status(fiber.StatusUnauthorized).SendString("out-of-turn endpoint access: complete the previous step of the signup process")
	}

	var signupSessionData map[string]any

	if err := json.Unmarshal([]byte(ssStr), &signupSessionData); err != nil {
		log.Println("ssm.go: ValidateSession: json.Unmarshal:", err)
		return fiber.ErrInternalServerError
	}

	c.Locals("signup_sess_data", signupSessionData)

	return c.Next()
}
