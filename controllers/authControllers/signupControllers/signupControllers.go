package signupControllers

import (
	"context"
	"i9chat/appGlobals"
	"i9chat/appTypes"
	"i9chat/services/auth/signupService"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
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

	sess.Set("signup_session", sessionData)

	if err := sess.Save(); err != nil {
		log.Println("signupControllers.go: RequestNewAccount: sess.Save:", err)
		return fiber.ErrInternalServerError
	}

	return c.JSON(respData)
}

func VerifyEmail(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sess := c.Locals("session").(*session.Session)

	signupSession := sess.Get("signup_session").(*appTypes.SignupSession)

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

	sess.Set("signup_session", updatedSessionData)

	if err := sess.Save(); err != nil {
		log.Println("signupControllers.go: VerifyEmail: sess.Save:", err)
		return fiber.ErrInternalServerError
	}

	return c.JSON(respData)
}

func RegisterUser(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sess := c.Locals("session").(*session.Session)

	signupSession := sess.Get("signup_session").(*appTypes.SignupSession)

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
