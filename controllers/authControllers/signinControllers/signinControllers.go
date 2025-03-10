package signinControllers

import (
	"context"
	"encoding/json"
	"i9chat/services/auth/signinService"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

func Signin(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var body signInBody

	body_err := c.BodyParser(&body)
	if body_err != nil {
		return body_err
	}

	if val_err := body.Validate(); val_err != nil {
		return val_err
	}

	respData, authJwt, app_err := signinService.Signin(ctx, body.EmailOrUsername, body.Password)
	if app_err != nil {
		return app_err
	}

	usd, err := json.Marshal(map[string]any{"authJwt": authJwt})
	if err != nil {
		log.Println("signinControllers.go: Signin: json.Marshal:", err)
		return fiber.ErrInternalServerError
	}

	c.Cookie(&fiber.Cookie{
		Name:     "user",
		Value:    string(usd),
		Path:     "/api/app",
		MaxAge:   int(10 * 24 * time.Hour / time.Second),
		HTTPOnly: true,
		Secure:   false,
	})

	return c.JSON(respData)
}
