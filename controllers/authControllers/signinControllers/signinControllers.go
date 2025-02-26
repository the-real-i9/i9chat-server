package signinControllers

import (
	"context"
	"i9chat/appGlobals"
	"i9chat/services/auth/signinService"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

func Signin(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	sess, err := appGlobals.SessionStore.Get(c)
	if err != nil {
		log.Println("signinControllers.go: Signin: SessionStore.Get:", err)
		return fiber.ErrInternalServerError
	}

	sess.Set("user", map[string]any{"authJwt": authJwt})
	sess.SetExpiry(10 * (24 * time.Hour))
	appGlobals.SessionStore.CookiePath = "/api/app"

	if err := sess.Save(); err != nil {
		log.Println("signinControllers.go: Signin: sess.Save:", err)
		return fiber.ErrInternalServerError
	}

	return c.JSON(respData)
}
