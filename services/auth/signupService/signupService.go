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

func RequestNewAccount(ctx context.Context, email string) (any, error) {
	accExists, err := appModel.AccountExists(ctx, email)
	if err != nil {
		return "", err
	}

	if accExists {
		return "", fiber.NewError(fiber.StatusBadRequest, "signup error: an account with", email, "already exists")
	}

	verfCode, expires := securityServices.GenerateVerifCodeExp()

	go mailService.SendMail(email, "Email Verification", fmt.Sprintf("Your email verification code is: <b>%d</b>", verfCode))

	sessionId, err := appModel.NewSignupSession(ctx, email, verfCode)
	if err != nil {
		return "", err
	}

	signupSessionJwt, err := securityServices.JwtSign(appTypes.SignupSessionData{
		SessionId: sessionId,
		Email:     email,
		Step:      "verify email",
	}, os.Getenv("SIGNUP_SESSION_JWT_SECRET"), expires)

	if err != nil {
		return "", err
	}

	respData := map[string]any{
		"msg":           "A 6-digit verification code has been sent to " + email,
		"session_token": signupSessionJwt,
	}

	return respData, nil
}

func VerifyEmail(ctx context.Context, sessionId string, inputVerfCode int, email string) (any, error) {
	isSuccess, err := appModel.VerifyEmail(ctx, sessionId, inputVerfCode)
	if err != nil {
		return "", err
	}

	if !isSuccess {
		return "", fiber.NewError(fiber.StatusBadRequest, "email verification error: incorrect verification code")
	}

	go mailService.SendMail(email, "Email Verification Success", fmt.Sprintf("Your email %s has been verified!", email))

	signupSessionJwt, err := securityServices.JwtSign(appTypes.SignupSessionData{
		SessionId: sessionId,
		Email:     email,
		Step:      "register user",
	}, os.Getenv("SIGNUP_SESSION_JWT_SECRET"), time.Now().UTC().Add(1*time.Hour))

	if err != nil {
		return "", err
	}

	respData := map[string]any{
		"msg":           fmt.Sprintf("Your email '%s' has been verified!", email),
		"session_token": signupSessionJwt,
	}

	return respData, nil
}

func RegisterUser(ctx context.Context, sessionId, email, username, password, geolocation string) (any, error) {
	accExists, err := appModel.AccountExists(ctx, username)
	if err != nil {
		return nil, err
	}

	if accExists {
		return nil, fiber.NewError(fiber.StatusBadRequest, "username error: username", username, "is unavailable")
	}

	hashedPassword, err := securityServices.HashPassword(password)
	if err != nil {
		return nil, err
	}

	newUser, err := user.New(ctx, email, username, hashedPassword, geolocation)
	if err != nil {
		return nil, err
	}

	authJwt, err := securityServices.JwtSign(appTypes.ClientUser{
		Id:       newUser.Id,
		Username: newUser.Username,
	}, os.Getenv("AUTH_JWT_SECRET"), time.Now().UTC().Add(365*24*time.Hour)) // 1 year

	if err != nil {
		return "", err
	}

	go appModel.EndSignupSession(sessionId)

	respData := map[string]any{
		"msg":     "Signup success!",
		"user":    newUser,
		"authJwt": authJwt,
	}

	return respData, nil
}
