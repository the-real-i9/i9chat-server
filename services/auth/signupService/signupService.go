package signupService

import (
	"context"
	"fmt"
	"i9chat/appTypes"
	"i9chat/helpers"
	user "i9chat/models/userModel"
	"i9chat/services/mailService"
	"i9chat/services/securityServices"
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
		return nil, nil, fiber.NewError(fiber.StatusBadRequest, "signup error: an account with", email, "already exists")
	}

	verfCode, expires := securityServices.GenerateVerifCodeExp()

	go mailService.SendMail(email, "Email Verification", fmt.Sprintf("Your email verification code is: <b>%d</b>", verfCode))

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

func VerifyEmail(ctx context.Context, sessionData map[string]any, inputVerfCode int) (any, map[string]any, error) {
	var sd struct {
		Email        string
		VCode        int
		VCodeExpires time.Time
	}

	helpers.MapToStruct(sessionData, &sd)

	if sd.VCode != inputVerfCode {
		return "", nil, fiber.NewError(fiber.StatusBadRequest, "email verification error: incorrect verification code")
	}

	if sd.VCodeExpires.Before(time.Now()) {
		return "", nil, fiber.NewError(fiber.StatusBadRequest, "email verification error: verification code expired")
	}

	go mailService.SendMail(sd.Email, "Email Verification Success", fmt.Sprintf("Your email %s has been verified!", sd.Email))

	newSessionData := map[string]any{"email": sd.Email}

	respData := map[string]any{
		"msg": fmt.Sprintf("Your email '%s' has been verified!", sd.Email),
	}

	return respData, newSessionData, nil
}

func RegisterUser(ctx context.Context, sessionData map[string]any, username, password, phone string, geolocation appTypes.UserGeolocation) (any, string, error) {
	email := sessionData["email"].(string)

	userExists, err := user.Exists(ctx, username)
	if err != nil {
		return nil, "", err
	}

	if userExists {
		return nil, "", fiber.NewError(fiber.StatusBadRequest, "signup error: username", username, "is unavailable")
	}

	hashedPassword, err := securityServices.HashPassword(password)
	if err != nil {
		return nil, "", err
	}

	newUser, err := user.New(ctx, email, username, phone, hashedPassword, geolocation)
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
