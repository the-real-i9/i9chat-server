package signinControllers

import (
	"context"
	"i9chat/services/auth/signinService"

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
		return c.Status(fiber.StatusBadRequest).SendString(val_err.Error())
	}

	respData, app_err := signinService.Signin(ctx, body.EmailOrUsername, body.Password)
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}
