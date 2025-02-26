package signupControllers

import (
	"context"
	"encoding/json"
	"i9chat/appGlobals"
	"i9chat/services/auth/signupService"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

func RequestNewAccount(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	sess, err := appGlobals.SessionStore.Get(c)
	if err != nil {
		log.Println("signupControllers.go: RequestNewAccount: SessionStore.Get:", err)
		return fiber.ErrInternalServerError
	}

	sd, err := json.Marshal(sessionData)
	if err != nil {
		log.Println("signupControllers.go: RequestNewAccount: json.Marshal:", err)
		return fiber.ErrInternalServerError
	}

	sess.Set("signup", sd)
	sess.SetExpiry(time.Hour)
	appGlobals.SessionStore.CookiePath = "/api/auth/signup/verify_email"

	if err := sess.Save(); err != nil {
		log.Println("signupControllers.go: RequestNewAccount: sess.Save:", err)
		return fiber.ErrInternalServerError
	}

	return c.JSON(respData)
}

func VerifyEmail(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sessionData := c.Locals("signup_sess_data").(map[string]any)

	var body verifyEmailBody

	body_err := c.BodyParser(&body)
	if body_err != nil {
		return body_err
	}

	if val_err := body.Validate(); val_err != nil {
		return val_err
	}

	respData, updatedSession, app_err := signupService.VerifyEmail(ctx, sessionData, body.Code)
	if app_err != nil {
		return app_err
	}

	sess, err := appGlobals.SessionStore.Get(c)
	if err != nil {
		log.Println("signupControllers.go: VerifyEmail: SessionStore.Get:", err)
		return fiber.ErrInternalServerError
	}

	usd, err := json.Marshal(updatedSession)
	if err != nil {
		log.Println("signupControllers.go: VerifyEmail: json.Marshal:", err)
		return fiber.ErrInternalServerError
	}

	sess.Set("signup", usd)
	sess.SetExpiry(time.Hour)
	appGlobals.SessionStore.CookiePath = "/api/auth/signup/register_user"

	if err := sess.Save(); err != nil {
		log.Println("signupControllers.go: VerifyEmail: sess.Save:", err)
		return fiber.ErrInternalServerError
	}

	return c.JSON(respData)
}

func RegisterUser(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	respData, authJwt, app_err := signupService.RegisterUser(ctx, sessionData, body.Username, body.Password, body.Phone, body.Geolocation)
	if app_err != nil {
		return app_err
	}

	sess, err := appGlobals.SessionStore.Get(c)
	if err != nil {
		log.Println("signupControllers.go: RegisterUser: SessionStore.Get:", err)
		return fiber.ErrInternalServerError
	}

	sess.Delete("signup")
	sess.Set("user", map[string]any{"authJwt": authJwt})
	sess.SetExpiry(10 * (24 * time.Hour))
	appGlobals.SessionStore.CookiePath = "/api/app"

	if err := sess.Save(); err != nil {
		log.Println("signupControllers.go: RegisterUser: sess.Save:", err)
		return fiber.ErrInternalServerError
	}

	return c.JSON(respData)
}
