package passwordResetControllers

import (
	"i9chat/src/helpers"
	"i9chat/src/services/auth/passwordResetService"
	"log"
	"time"

	"github.com/goccy/go-json"

	"github.com/gofiber/fiber/v2"
)

// Forgot Password - Request Password Reset
//
//	@Summary		Password Reset - Step 1
//	@Description	Submit your email to request a password reset
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//
//	@Param			email	body		string									true	"Provide your email address"
//
//	@Success		200		{object}	passwordResetService.passReset1RespT	"Proceed to email confirmation"
//	@Header			200		{array}		Set-Cookie								"Password Reset session response cookie"
//
//	@Failure		404		{object}	appErrors.HTTPError						"No user with this email exists"
//
//	@Failure		500		{object}	appErrors.HTTPError
//
//	@Router			/auth/forgot_password/request_password_reset [post]
func RequestPasswordReset(c *fiber.Ctx) error {
	ctx := c.Context()

	var body requestPasswordResetBody

	body_err := c.BodyParser(&body)
	if body_err != nil {
		return body_err
	}

	if val_err := body.Validate(); val_err != nil {
		return val_err
	}

	respData, sessionData, app_err := passwordResetService.RequestPasswordReset(ctx, body.Email)
	if app_err != nil {
		return app_err
	}

	sd, err := json.Marshal(sessionData)
	if err != nil {
		log.Println("passwordResetControllers.go: RequestPasswordReset: json.Marshal:", err)
		return fiber.ErrInternalServerError
	}

	c.Cookie(helpers.Cookie("passwordReset", string(sd), int(time.Hour/time.Second)))

	return c.JSON(respData)
}

// Forgot Password - Confirm Email
//
//	@Summary		Password Reset - Step 2
//	@Description	Provide the 6-digit token sent to email
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//
//	@Param			token	body		string									true	"6-digit token"
//	@Param			Cookie	header		[]string								true	"Password Reset session request cookie"
//
//	@Success		200		{object}	passwordResetService.passReset2RespT	"Email confirmed. You're about to reset your password"
//	@Header			200		{array}		Set-Cookie								"Password Reset session request cookie"
//
//	@Failure		400		{object}	appErrors.HTTPError						"Incorrect or expired confirmation token"
//	@Header			400		{array}		Set-Cookie								"Password Reset session request cookie"
//
//	@Failure		500		{object}	appErrors.HTTPError
//	@Header			500		{array}		Set-Cookie	"Password Reset session request cookie"
//
//	@Router			/auth/forgot_password/confirm_email [post]
func ConfirmEmail(c *fiber.Ctx) error {
	ctx := c.Context()

	sessionData := c.Locals("passwordReset_sess_data").(map[string]any)

	var body confirmEmailBody

	body_err := c.BodyParser(&body)
	if body_err != nil {
		return body_err
	}

	if val_err := body.Validate(); val_err != nil {
		return val_err
	}

	respData, newSessionData, app_err := passwordResetService.ConfirmEmail(ctx, sessionData, body.Token)
	if app_err != nil {
		return app_err
	}

	nsd, err := json.Marshal(newSessionData)
	if err != nil {
		log.Println("passwordResetControllers.go: ConfirmEmail: json.Marshal:", err)
		return fiber.ErrInternalServerError
	}

	c.Cookie(helpers.Cookie("passwordReset", string(nsd), int(time.Hour/time.Second)))

	return c.JSON(respData)
}

// Forgot Password - Reset Password
//
//	@Summary		Password Reset user - Step 3
//	@Description	Set new password
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//
//	@Param			newPassword			body		string									true	"Choose a new password"
//	@Param			confirmNewPassword	body		string									true	"Conform new password"
//
//	@Param			Cookie				header		[]string								true	"Password Reset session request cookie"
//
//	@Success		200					{object}	passwordResetService.passReset3RespT	"Password changed successfully"
//
//	@Failure		400					{object}	appErrors.HTTPError						"Passwords mismatch."
//	@Header			400					{array}		Set-Cookie								"Password Reset session response cookie"
//
//	@Failure		500					{object}	appErrors.HTTPError
//	@Header			500					{array}		Set-Cookie	"Password Reset session response cookie"
//
//	@Router			/auth/forgot_password/reset_password [post]
func ResetPassword(c *fiber.Ctx) error {
	ctx := c.Context()

	sessionData := c.Locals("passwordReset_sess_data").(map[string]any)

	var body resetPasswordBody

	body_err := c.BodyParser(&body)
	if body_err != nil {
		return body_err
	}

	if val_err := body.Validate(); val_err != nil {
		log.Println(val_err)
		return val_err
	}

	respData, app_err := passwordResetService.ResetPassword(ctx, sessionData, body.NewPassword)
	if app_err != nil {
		return app_err
	}

	c.Cookie(helpers.Cookie("passwordReset", "", 0))

	return c.JSON(respData)
}
