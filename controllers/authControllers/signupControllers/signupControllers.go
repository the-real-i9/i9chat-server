package signupControllers

import (
	"context"
	"i9chat/appGlobals"
	"i9chat/appTypes"
	"i9chat/services/auth/signupService"
	"log"

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
		return c.Status(fiber.StatusBadRequest).SendString(val_err.Error())
	}

	respData, sessionData, app_err := signupService.RequestNewAccount(ctx, body.Email)
	if app_err != nil {
		return app_err
	}

	sess, err := appGlobals.SignupSessionStore.Get(c)
	if err != nil {
		log.Println("signupControllers.go: RequestNewAccount: SignupSessionStore.Get:", err)
		return fiber.ErrInternalServerError
	}

	sess.Set("session", sessionData)

	if err := sess.Save(); err != nil {
		log.Println("signupControllers.go: RequestNewAccount: sess.Save:", err)
		return fiber.ErrInternalServerError
	}

	return c.JSON(respData)
}

func VerifyEmail(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sess, err := appGlobals.SignupSessionStore.Get(c)
	if err != nil {
		log.Println("signupControllers.go: VerifyEmail: SignupSessionStore.Get:", err)
		return fiber.ErrInternalServerError
	}

	signupSession := sess.Get("session").(*appTypes.SignupSession)

	if signupSession.Step != "verify email" {
		return c.Status(fiber.StatusUnauthorized).SendString("invalid session cookie at endpoint")
	}

	var body verifyEmailBody

	body_err := c.BodyParser(&body)
	if body_err != nil {
		return body_err
	}

	if val_err := body.Validate(); val_err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(val_err.Error())
	}

	respData, updatedSessionData, app_err := signupService.VerifyEmail(ctx, signupSession.Data, body.Code)
	if app_err != nil {
		return app_err
	}

	sess.Set("session", updatedSessionData)

	if err := sess.Save(); err != nil {
		log.Println("signupControllers.go: VerifyEmail: sess.Save:", err)
		return fiber.ErrInternalServerError
	}

	return c.JSON(respData)
}

func RegisterUser(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sess, err := appGlobals.SignupSessionStore.Get(c)
	if err != nil {
		log.Println("signupControllers.go: VerifyEmail: SignupSessionStore.Get:", err)
		return fiber.ErrInternalServerError
	}

	signupSession := sess.Get("session").(*appTypes.SignupSession)

	if signupSession.Step != "verify email" {
		return c.Status(fiber.StatusUnauthorized).SendString("invalid session cookie at endpoint")
	}

	var body registerUserBody

	body_err := c.BodyParser(&body)
	if body_err != nil {
		return body_err
	}

	if val_err := body.Validate(); val_err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(val_err.Error())
	}

	respData, app_err := signupService.RegisterUser(ctx, signupSession.Data, body.Username, body.Password, body.Geolocation)
	if app_err != nil {
		return app_err
	}

	if err := sess.Destroy(); err != nil {
		log.Println("signupControllers.go: RegisterUser: sess.Destroy:", err)
		return fiber.ErrInternalServerError
	}

	return c.JSON(respData)
}
