package sessionMiddlewares

import (
	"log"

	"github.com/goccy/go-json"

	"github.com/gofiber/fiber/v2"
)

func PasswordResetSession(c *fiber.Ctx) error {
	prsStr := c.Cookies("passwordReset")

	if prsStr == "" {
		return c.Status(fiber.StatusUnauthorized).SendString("out-of-turn endpoint access: complete the previous step of the password reset process")
	}

	var passwordResetSessionData map[string]any

	if err := json.Unmarshal([]byte(prsStr), &passwordResetSessionData); err != nil {
		log.Println("prsm.go: PasswordResetSession: json.Unmarshal:", err)
		return fiber.ErrInternalServerError
	}

	c.Locals("passwordReset_sess_data", passwordResetSessionData)

	return c.Next()
}
