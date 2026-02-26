package passwordResetControllers

import (
	"i9chat/src/helpers"
	"i9chat/src/services/auth/passwordResetService"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/vmihailenco/msgpack/v5"
)

func RequestPasswordReset(c fiber.Ctx) error {
	ctx := c.Context()

	var body requestPasswordResetBody

	err := c.Bind().MsgPack(&body)
	if err != nil {
		return err
	}

	if err := body.Validate(); err != nil {
		return err
	}

	respData, sessionData, err := passwordResetService.RequestPasswordReset(ctx, body.Email)
	if err != nil {
		return err
	}

	reqSession := map[string]any{
		"passwordReset": sessionData,
	}

	c.Cookie(helpers.Session(reqSession, "/api/auth/forgot_password/confirm_email", int(time.Hour/time.Second)))

	return c.MsgPack(respData)
}

func ConfirmEmail(c fiber.Ctx) error {
	ctx := c.Context()

	sessionData := c.Locals("passwordReset_sess_data").(msgpack.RawMessage)

	var body confirmEmailBody

	err := c.Bind().MsgPack(&body)
	if err != nil {
		return err
	}

	if err := body.Validate(); err != nil {
		return err
	}

	respData, newSessionData, err := passwordResetService.ConfirmEmail(ctx, sessionData, body.Token)
	if err != nil {
		return err
	}

	reqSession := map[string]any{
		"passwordReset": newSessionData,
	}

	c.Cookie(helpers.Session(reqSession, "/api/auth/forgot_password/reset_password", int(time.Hour/time.Second)))

	return c.MsgPack(respData)
}

func ResetPassword(c fiber.Ctx) error {
	ctx := c.Context()

	sessionData := c.Locals("passwordReset_sess_data").(msgpack.RawMessage)

	var body resetPasswordBody

	err := c.Bind().MsgPack(&body)
	if err != nil {
		return err
	}

	if err := body.Validate(); err != nil {
		helpers.LogError(err)
		return err
	}

	respData, err := passwordResetService.ResetPassword(ctx, sessionData, body.NewPassword)
	if err != nil {
		return err
	}

	c.ClearCookie()

	return c.MsgPack(respData)
}
