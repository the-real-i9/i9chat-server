package middlewares

import (
	"fmt"
	"os"
	"utils/helpers"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

func CheckAccountRequested(c *websocket.Conn) (map[string]any, error) {
	token := c.Headers("Authorization")

	if token == "" {
		return nil, fmt.Errorf("signup error: no ongoing signup session. you must first submit your email and attach the autorization token sent")
	}

	sessionData, err := helpers.JwtParse(token, os.Getenv("SIGNUP_JWT_SECRET"))
	if err != nil {
		if err.Error() == "authentication error: invalid jwt" {
			return nil, fmt.Errorf("signup error: invalid signup session token")
		}
		if err.Error() == "authentication error: jwt expired" {
			return nil, fmt.Errorf("signup error: signup session expired")
		}
	}

	return sessionData, nil
}

func CheckEmailVerified(c *fiber.Ctx) error {
	ber := c.Get("Authorization")

	// In addition to the above:
	// If verified is false, reply: "Email not verified."

	fmt.Println(ber)

	return c.SendStatus(200)
}
