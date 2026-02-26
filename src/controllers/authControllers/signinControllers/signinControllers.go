package signinControllers

import (
	"i9chat/src/helpers"
	"i9chat/src/services/auth/signinService"
	"time"

	"github.com/gofiber/fiber/v3"
)

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
