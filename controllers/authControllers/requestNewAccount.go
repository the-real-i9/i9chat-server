package authControllers

import (
	"fmt"
	"i9chat/appTypes"
	"i9chat/models/appModel"
	"i9chat/services/appServices"
	"i9chat/services/authServices"
	"os"
)

func requestNewAccount(email string) (string, error) {
	accExists, err := appModel.AccountExists(email)
	if err != nil {
		return "", err
	}

	if accExists {
		return "", fmt.Errorf("signup error: an account with '%s' already exists", email)
	}

	verfCode, expires := authServices.GenerateVerifCodeExp()

	go appServices.SendMail(email, "Email Verification", fmt.Sprintf("Your email verification code is: <b>%d</b>", verfCode))

	sessionId, err := appModel.NewSignupSession(email, verfCode)
	if err != nil {
		return "", err
	}

	signupSessionJwt := authServices.JwtSign(appTypes.SignupSessionData{
		SessionId: sessionId,
		Email:     email,
		Step:      "verify email",
	}, os.Getenv("SIGNUP_SESSION_JWT_SECRET"), expires)

	return signupSessionJwt, nil
}
