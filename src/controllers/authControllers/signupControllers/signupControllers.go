package signupControllers

import (
	"encoding/json"
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

	sd, err := json.Marshal(sessionData)
	if err != nil {
		log.Println("signupControllers.go: RequestNewAccount: json.Marshal:", err)
		return fiber.ErrInternalServerError
	}

	c.Cookie(helpers.Cookie("signup", string(sd), int(time.Hour/time.Second)))
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

	nsd, err := json.Marshal(newSessionData)
	if err != nil {
		log.Println("signupControllers.go: VerifyEmail: json.Marshal:", err)
		return fiber.ErrInternalServerError
	}

	c.Cookie(helpers.Cookie("signup", string(nsd), int(time.Hour/time.Second)))

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

	usd, err := json.Marshal(map[string]any{"authJwt": authJwt})
	if err != nil {
		log.Println("signupControllers.go: RegisterUser: json.Marshal:", err)
		return fiber.ErrInternalServerError
	}

	c.Cookie(helpers.Cookie("user", string(usd), int(10*24*time.Hour/time.Second)))
	c.Cookie(helpers.Cookie("signup", "", 0))

	return c.JSON(respData)
}
