package signinControllers

import (
	"i9chat/src/helpers"
	"i9chat/src/services/auth/signinService"
	"time"

	"github.com/gofiber/fiber/v3"
)

// Signin
//
//	@Summary		Signin user
//	@Description	Signin with email/username and password
//	@Tags			auth
//	@Accept			application/vnd.msgpack
//	@Produce		application/vnd.msgpack
//
//	@Param			emailOrUsername	body		string						true	"Email or Username"
//	@Param			password		body		string						true	"User Password"
//
//	@Success		200				{object}	signinService.signinRespT	"Signin Success!"
//	@Header			200				{string}	Set-Cookie					"Authenticated user session response cookie containing auth JWT"
//
//	@Failure		400				{object}	appErrors.HTTPError
//
//	@Failure		404				{object}	appErrors.HTTPError	"Incorrect credentials"
//
//	@Failure		500				{object}	appErrors.HTTPError
//
//	@Router			/auth/signin [post]
func Signin(c fiber.Ctx) error {
	ctx := c.Context()

	var body signInBody

	err := c.Bind().MsgPack(&body)
	if err != nil {
		return err
	}

	if err := body.Validate(); err != nil {
		return err
	}

	respData, authJwt, err := signinService.Signin(ctx, body.EmailOrUsername, body.Password)
	if err != nil {
		return err
	}

	reqSession := map[string]any{
		"user": map[string]any{"authJwt": authJwt},
	}

	c.Cookie(helpers.Session(reqSession, "/api/app", int(10*24*time.Hour/time.Second)))

	return c.MsgPack(respData)
}
