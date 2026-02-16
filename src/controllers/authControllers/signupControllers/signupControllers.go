package signupControllers

import (
	"i9chat/src/helpers"
	"i9chat/src/services/auth/signupService"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/vmihailenco/msgpack/v5"
)

// Signup - Request New Account
//
//	@Summary		Signup user - Step 1
//	@Description	Submit email to request a new account
//	@Tags			auth
//	@Accept			application/vnd.msgpack
//	@Produce		application/vnd.msgpack
//
//	@Param			email	body		string						true	"Provide your email address"
//
//	@Success		200		{object}	signupService.signup1RespT	"Proceed to email verification"
//	@Header			200		{string}	Set-cookie					"Signup session response cookie"
//
//	@Failure		400		{object}	appErrors.HTTPError			"An account with email already exists"
//
//	@Failure		500		{object}	appErrors.HTTPError
//
//	@Router			/auth/signup/request_new_account [post]

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

// Signup - Verify Email
//
//	@Summary		Signup user - Step 2
//	@Description	Provide the 6-digit code sent to email
//	@Tags			auth
//	@Accept			application/vnd.msgpack
//	@Produce		application/vnd.msgpack
//
//	@Param			code	body		string						true	"6-digit code"
//	@Param			Cookie	header		string						true	"Signup session request cookie"
//
//	@Success		200		{object}	signupService.signup2RespT	"Email verified"
//	@Header			200		{string}	Set-cookie					"Signup session response cookie"
//
//	@Failure		400		{object}	appErrors.HTTPError			"Incorrect or expired verification code"
//	@Header			400		{string}	Set-cookie					"Signup session response cookie"
//
//	@Failure		500		{object}	appErrors.HTTPError
//	@Header			500		{string}	Set-cookie	"Signup session response cookie"
//
//	@Router			/auth/signup/verify_email [post]
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

// Signup - Register user
//
//	@Summary		Signup user - Step 3
//	@Description	Provide remaining user credentials
//	@Tags			auth
//	@Accept			application/vnd.msgpack
//	@Produce		application/vnd.msgpack
//
//	@Param			username	body		string						true	"Choose a username"
//	@Param			password	body		string						true	"Choose a password"
//	@Param			bio			body		string						false	"User bio (optional)"
//
//	@Param			Cookie		header		string						true	"Signup session request cookie"
//
//	@Success		200			{object}	signupService.signup3RespT	"Signup Success"
//	@Header			200			{string}	Set-cookie					"Authenticated user session response cookie containing auth JWT"
//
//	@Failure		400			{object}	appErrors.HTTPError			"Incorrect or expired verification code"
//	@Header			400			{string}	Set-cookie					"Signup session response cookie"
//
//	@Failure		500			{object}	appErrors.HTTPError
//	@Header			500			{string}	Set-cookie	"Signup session response cookie"
//
//	@Router			/auth/signup/register_user [post]
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

	return c.Status(fiber.StatusCreated).JSON(respData)
}
