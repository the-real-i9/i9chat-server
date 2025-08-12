package signupService

import (
	"context"
	"fmt"
	"i9chat/src/appTypes"
	"i9chat/src/helpers"
	user "i9chat/src/models/userModel"
	"i9chat/src/services/mailService"
	"i9chat/src/services/securityServices"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
)

func RequestNewAccount(ctx context.Context, email string) (any, map[string]any, error) {

	userExists, err := user.Exists(ctx, email)
	if err != nil {
		return nil, nil, err
	}

	if userExists {
		return nil, nil, fiber.NewError(fiber.StatusConflict, fmt.Sprintf("An account with %s already exists", email))
	}

	verfCode, expires := securityServices.GenerateTokenCodeExp()

	go mailService.SendMail(email, "Email Verification", fmt.Sprintf("Your email verification code is: <b>%s</b>", verfCode))

	sessionData := map[string]any{
		"email":        email,
		"vCode":        verfCode,
		"vCodeExpires": expires,
	}

	respData := map[string]any{
		"msg": "A 6-digit verification code has been sent to " + email,
	}

	return respData, sessionData, nil
}

func VerifyEmail(ctx context.Context, sessionData map[string]any, inputVerfCode string) (any, map[string]any, error) {
	var sd struct {
		Email        string
		VCode        string
		VCodeExpires time.Time
	}

	helpers.ToStruct(sessionData, &sd)

	if sd.VCode != inputVerfCode {
		return "", nil, fiber.NewError(fiber.StatusBadRequest, "Incorrect verification code! Check or Re-submit your email.")
	}

	if sd.VCodeExpires.Before(time.Now()) {
		return "", nil, fiber.NewError(fiber.StatusBadRequest, "Verification code expired! Re-submit your email.")
	}

	go mailService.SendMail(sd.Email, "Email Verification Success", fmt.Sprintf("Your email %s has been verified!", sd.Email))

	newSessionData := map[string]any{"email": sd.Email}

	respData := map[string]any{
		"msg": fmt.Sprintf("Your email '%s' has been verified!", sd.Email),
	}

	return respData, newSessionData, nil
}

func RegisterUser(ctx context.Context, sessionData map[string]any, username, password string) (any, string, error) {
	email := sessionData["email"].(string)

	userExists, err := user.Exists(ctx, username)
	if err != nil {
		return nil, "", err
	}

	if userExists {
		return nil, "", fiber.NewError(fiber.StatusConflict, "Username unavailable")
	}

	hashedPassword, err := securityServices.HashPassword(password)
	if err != nil {
		return nil, "", err
	}

	newUser, err := user.New(ctx, email, username, hashedPassword)
	if err != nil {
		return nil, "", err
	}

	authJwt, err := securityServices.JwtSign(appTypes.ClientUser{
		Username: username,
	}, os.Getenv("AUTH_JWT_SECRET"), time.Now().UTC().Add(10*24*time.Hour)) // 10 days

	if err != nil {
		return nil, "", err
	}

	respData := map[string]any{
		"msg":  "Signup success!",
		"user": newUser,
	}

	return respData, authJwt, nil
}
