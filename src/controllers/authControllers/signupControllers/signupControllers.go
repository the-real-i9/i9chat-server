package signupControllers

import (
	"i9chat/src/helpers"
	"i9chat/src/services/auth/signupService"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/vmihailenco/msgpack/v5"
)

func RequestNewAccount(c fiber.Ctx) error {
	ctx := c.Context()

	var body requestNewAccountBody

	err := c.Bind().MsgPack(&body)
	if err != nil {
		return err
	}

	if err := body.Validate(); err != nil {
		return err
	}

	respData, sessionData, err := signupService.RequestNewAccount(ctx, body.Email)
	if err != nil {
		return err
	}

	reqSession := map[string]any{
		"signup": sessionData,
	}

	c.Cookie(helpers.Session(reqSession, "/api/auth/signup/verify_email", int(time.Hour/time.Second)))

	return c.MsgPack(respData)
}

func VerifyEmail(c fiber.Ctx) error {
	ctx := c.Context()

	sessionData := c.Locals("signup_sess_data").(msgpack.RawMessage)

	var body verifyEmailBody

	err := c.Bind().MsgPack(&body)
	if err != nil {
		return err
	}

	if err := body.Validate(); err != nil {
		return err
	}

	respData, newSessionData, err := signupService.VerifyEmail(ctx, sessionData, body.Code)
	if err != nil {
		return err
	}

	reqSession := map[string]any{
		"signup": newSessionData,
	}

	c.Cookie(helpers.Session(reqSession, "/api/auth/signup/register_user", int(time.Hour/time.Second)))

	return c.MsgPack(respData)
}

func RegisterUser(c fiber.Ctx) error {
	ctx := c.Context()

	sessionData := c.Locals("signup_sess_data").(msgpack.RawMessage)

	var body registerUserBody

	err := c.Bind().MsgPack(&body)
	if err != nil {
		return err
	}

	if err := body.Validate(); err != nil {
		helpers.LogError(err)
		return err
	}

	respData, authJwt, err := signupService.RegisterUser(ctx, sessionData, body.Username, body.Password, body.Bio)
	if err != nil {
		return err
	}

	reqSession := map[string]any{
		"user": map[string]any{"authJwt": authJwt},
	}

	c.Cookie(helpers.Session(reqSession, "/api/app", int(10*24*time.Hour/time.Second)))

	return c.Status(fiber.StatusCreated).MsgPack(respData)
}
