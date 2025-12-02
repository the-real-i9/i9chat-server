package signinControllers

import (
	"i9chat/src/helpers"
	"i9chat/src/services/auth/signinService"
	"time"

	"github.com/gofiber/fiber/v2"
)

func Signin(c *fiber.Ctx) error {
	ctx := c.Context()

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

	c.Cookie(helpers.Cookie("user", helpers.ToJson(map[string]any{"authJwt": authJwt}), int(10*24*time.Hour/time.Second)))
	c.Cookie(helpers.Cookie("passwordReset", "", 0))

	return c.JSON(respData)
}
