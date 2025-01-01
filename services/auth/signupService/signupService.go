package signupService

import (
	"context"
	"fmt"
	"i9chat/appTypes"
	"i9chat/models/appModel"
	user "i9chat/models/userModel"
	"i9chat/services/mailService"
	"i9chat/services/securityServices"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
)

func RequestNewAccount(ctx context.Context, email string) (any, *appTypes.SignupSession, error) {
	accExists, err := appModel.AccountExists(ctx, email)
	if err != nil {
		return nil, nil, err
	}

	if accExists {
		return nil, nil, fiber.NewError(fiber.StatusBadRequest, "signup error: an account with", email, "already exists")
	}

	verfCode, expires := securityServices.GenerateVerifCodeExp()

	go mailService.SendMail(email, "Email Verification", fmt.Sprintf("Your email verification code is: <b>%d</b>", verfCode))

	sessionData := &appTypes.SignupSession{
		Step: "verify email",
		Data: &appTypes.SignupSessionData{Email: email, VerificationCode: verfCode, VerificationCodeExpires: expires},
	}

	respData := map[string]any{
		"msg": "A 6-digit verification code has been sent to " + email,
	}

	return respData, sessionData, nil
}

func VerifyEmail(ctx context.Context, sessionData *appTypes.SignupSessionData, inputVerfCode int) (any, *appTypes.SignupSession, error) {
	if sessionData.VerificationCode != inputVerfCode {
		return "", nil, fiber.NewError(fiber.StatusBadRequest, "email verification error: incorrect verification code")
	}

	if sessionData.VerificationCodeExpires.Before(time.Now()) {
		return "", nil, fiber.NewError(fiber.StatusBadRequest, "email verification error: verification code expired")
	}

	go mailService.SendMail(sessionData.Email, "Email Verification Success", fmt.Sprintf("Your email %s has been verified!", sessionData.Email))

	updatedSessionData := &appTypes.SignupSession{
		Step: "register user",
		Data: &appTypes.SignupSessionData{Email: sessionData.Email},
	}

	respData := map[string]any{
		"msg": fmt.Sprintf("Your email '%s' has been verified!", sessionData.Email),
	}

	return respData, updatedSessionData, nil
}

func RegisterUser(ctx context.Context, sessionData *appTypes.SignupSessionData, username, password, geolocation string) (any, error) {
	accExists, err := appModel.AccountExists(ctx, username)
	if err != nil {
		return nil, err
	}

	if accExists {
		return nil, fiber.NewError(fiber.StatusBadRequest, "signup error: username", username, "is unavailable")
	}

	hashedPassword, err := securityServices.HashPassword(password)
	if err != nil {
		return nil, err
	}

	newUser, err := user.New(ctx, sessionData.Email, username, hashedPassword, geolocation)
	if err != nil {
		return nil, err
	}

	authJwt, err := securityServices.JwtSign(appTypes.ClientUser{
		Id:       newUser.Id,
		Username: newUser.Username,
	}, os.Getenv("AUTH_JWT_SECRET"), time.Now().UTC().Add(10*24*time.Hour)) // 1 year

	if err != nil {
		return "", err
	}

	respData := map[string]any{
		"msg":     "Signup success!",
		"user":    newUser,
		"authJwt": authJwt,
	}

	return respData, nil
}
