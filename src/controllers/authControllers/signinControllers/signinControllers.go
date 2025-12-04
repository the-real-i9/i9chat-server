package signinControllers

import (
	"i9chat/src/helpers"
	"i9chat/src/services/auth/signinService"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Signin
//
//	@Summary		Signin user
//	@Description	Signin with email/username and password
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//
//	@Param			EmailOrUsername	body		string						true	"Email or Username"
//	@Param			Password		body		string						true	"User Password"
//
//	@Success		200				{object}	signinService.signinRespT	"Signin Success!"
//	@Header			200				{array}		Set-Cookie					"User session response cookie containing auth JWT"
//
//	@Failure		400				{object}	appErrors.HTTPError
//
//	@Failure		404				{object}	appErrors.HTTPError	"Incorrect credentials"
//
//	@Failure		500				{object}	appErrors.HTTPError
//
//	@Router			/auth/signin [post]
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
