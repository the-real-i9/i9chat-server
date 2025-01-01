package signupControllers

import (
	"context"
	"encoding/json"
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

	sd, err := json.Marshal(sessionData)
	if err != nil {
		log.Println("signupControllers.go: RequestNewAccount: json.Marshal:", err)
		return fiber.ErrInternalServerError
	}

	sess.Set("signup_session", sd)

	if err := sess.Save(); err != nil {
		log.Println("signupControllers.go: RequestNewAccount: sess.Save:", err)
		return fiber.ErrInternalServerError
	}

	return c.JSON(respData)
}

func VerifyEmail(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sessionData := c.Locals("signup_session_data").(*appTypes.SignupSessionData)

	var body verifyEmailBody

	body_err := c.BodyParser(&body)
	if body_err != nil {
		return body_err
	}

	if val_err := body.Validate(); val_err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(val_err.Error())
	}

	respData, updatedSessionData, app_err := signupService.VerifyEmail(ctx, sessionData, body.Code)
	if app_err != nil {
		return app_err
	}

	sess, err := appGlobals.SignupSessionStore.Get(c)
	if err != nil {
		log.Println("signupControllers.go: VerifyEmail: SignupSessionStore.Get:", err)
		return fiber.ErrInternalServerError
	}

	usd, err := json.Marshal(updatedSessionData)
	if err != nil {
		log.Println("signupControllers.go: RequestNewAccount: json.Marshal:", err)
		return fiber.ErrInternalServerError
	}

	sess.Set("signup_session", usd)

	if err := sess.Save(); err != nil {
		log.Println("signupControllers.go: VerifyEmail: sess.Save:", err)
		return fiber.ErrInternalServerError
	}

	return c.JSON(respData)
}

func RegisterUser(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sessionData := c.Locals("signup_session_data").(*appTypes.SignupSessionData)

	var body registerUserBody

	body_err := c.BodyParser(&body)
	if body_err != nil {
		return body_err
	}

	if val_err := body.Validate(); val_err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(val_err.Error())
	}

	respData, app_err := signupService.RegisterUser(ctx, sessionData, body.Username, body.Password, body.Geolocation)
	if app_err != nil {
		return app_err
	}

	sess, err := appGlobals.SignupSessionStore.Get(c)
	if err != nil {
		log.Println("signupControllers.go: RegisterUser: SignupSessionStore.Get:", err)
		return fiber.ErrInternalServerError
	}

	if err := sess.Destroy(); err != nil {
		log.Println("signupControllers.go: RegisterUser: sess.Destroy:", err)
		return fiber.ErrInternalServerError
	}

	return c.JSON(respData)
}
