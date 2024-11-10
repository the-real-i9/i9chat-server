package authControllers

import (
	"fmt"
	"i9chat/appTypes"
	"i9chat/models/appModel"
	"i9chat/services/appServices"
	"i9chat/services/authServices"
	"os"
	"time"
)

func verifyEmail(sessionId string, inputVerfCode int, email string) (string, error) {
	isSuccess, err := appModel.VerifyEmail(sessionId, inputVerfCode)
	if err != nil {
		return "", err
	}

	if !isSuccess {
		return "", fmt.Errorf("email verification error: incorrect verification code")
	}

	go appServices.SendMail(email, "Email Verification Success", fmt.Sprintf("Your email %s has been verified!", email))

	signupSessionJwt := authServices.JwtSign(appTypes.SignupSessionData{
		SessionId: sessionId,
		Email:     email,
		Step:      "register user",
	}, os.Getenv("SIGNUP_SESSION_JWT_SECRET"), time.Now().UTC().Add(1*time.Hour))

	return signupSessionJwt, nil
}
