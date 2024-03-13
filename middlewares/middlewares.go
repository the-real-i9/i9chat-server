package middlewares

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func CheckAccountRequested(c *fiber.Ctx) error {
	ber := c.Get("Authorization")

	// If no token, reply: "No ongoing signup session. You might want to check if you've attached the required signup session token in the Authorization header.

	// If token is incorrect, reply: "Invalid signup session token"

	// If token is expired, reply: "Signup session expired"

	fmt.Println(ber)

	return c.SendStatus(200)
}

func CheckEmailVerified(c *fiber.Ctx) error {
	ber := c.Get("Authorization")

	// In addition to the above:
	// If verified is false, reply: "Email not verified."

	fmt.Println(ber)

	return c.SendStatus(200)
}
