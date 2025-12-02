package signupControllers

import (
	"i9chat/src/helpers"
	"i9chat/src/services/auth/signupService"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

func RequestNewAccount(c *fiber.Ctx) error {
	ctx := c.Context()

	var body requestNewAccountBody

	body_err := c.BodyParser(&body)
	if body_err != nil {
		return body_err
	}

	if val_err := body.Validate(); val_err != nil {
		return val_err
	}

	respData, sessionData, app_err := signupService.RequestNewAccount(ctx, body.Email)
	if app_err != nil {
		return app_err
	}

	c.Cookie(helpers.Cookie("signup", helpers.ToJson(sessionData), int(time.Hour/time.Second)))
	c.Cookie(helpers.Cookie("passwordReset", "", 0))

	return c.JSON(respData)
}

func VerifyEmail(c *fiber.Ctx) error {
	ctx := c.Context()

	sessionData := c.Locals("signup_sess_data").(map[string]any)

	var body verifyEmailBody

	body_err := c.BodyParser(&body)
	if body_err != nil {
		return body_err
	}

	if val_err := body.Validate(); val_err != nil {
		return val_err
	}

	respData, newSessionData, app_err := signupService.VerifyEmail(ctx, sessionData, body.Code)
	if app_err != nil {
		return app_err
	}

	c.Cookie(helpers.Cookie("signup", helpers.ToJson(newSessionData), int(time.Hour/time.Second)))

	return c.JSON(respData)
}

func RegisterUser(c *fiber.Ctx) error {
	ctx := c.Context()

	sessionData := c.Locals("signup_sess_data").(map[string]any)

	var body registerUserBody

	body_err := c.BodyParser(&body)
	if body_err != nil {
		return body_err
	}

	if val_err := body.Validate(); val_err != nil {
		log.Println(val_err)
		return val_err
	}

	respData, authJwt, app_err := signupService.RegisterUser(ctx, sessionData, body.Username, body.Password)
	if app_err != nil {
		return app_err
	}

	c.Cookie(helpers.Cookie("user", helpers.ToJson(map[string]any{"authJwt": authJwt}), int(10*24*time.Hour/time.Second)))
	c.Cookie(helpers.Cookie("signup", "", 0))

	return c.JSON(respData)
}
