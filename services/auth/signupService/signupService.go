package signupService

import (
	"fmt"
	"i9chat/appTypes"
	"i9chat/models/appModel"
	user "i9chat/models/userModel"
	"i9chat/services/mailService"
	"i9chat/services/securityServices"
	"os"
	"time"
)

func RequestNewAccount(email string) (any, error) {
	accExists, err := appModel.AccountExists(email)
	if err != nil {
		return "", err
	}

	if accExists {
		return "", fmt.Errorf("signup error: an account with '%s' already exists", email)
	}

	verfCode, expires := securityServices.GenerateVerifCodeExp()

	go mailService.SendMail(email, "Email Verification", fmt.Sprintf("Your email verification code is: <b>%d</b>", verfCode))

	sessionId, err := appModel.NewSignupSession(email, verfCode)
	if err != nil {
		return "", err
	}

	signupSessionJwt := securityServices.JwtSign(appTypes.SignupSessionData{
		SessionId: sessionId,
		Email:     email,
		Step:      "verify email",
	}, os.Getenv("SIGNUP_SESSION_JWT_SECRET"), expires)

	respData := map[string]any{
		"msg":           "A 6-digit verification code has been sent to " + email,
		"session_token": signupSessionJwt,
	}

	return respData, nil
}

func VerifyEmail(sessionId string, inputVerfCode int, email string) (any, error) {
	isSuccess, err := appModel.VerifyEmail(sessionId, inputVerfCode)
	if err != nil {
		return "", err
	}

	if !isSuccess {
		return "", fmt.Errorf("email verification error: incorrect verification code")
	}

	go mailService.SendMail(email, "Email Verification Success", fmt.Sprintf("Your email %s has been verified!", email))

	signupSessionJwt := securityServices.JwtSign(appTypes.SignupSessionData{
		SessionId: sessionId,
		Email:     email,
		Step:      "register user",
	}, os.Getenv("SIGNUP_SESSION_JWT_SECRET"), time.Now().UTC().Add(1*time.Hour))

	respData := map[string]any{
		"msg":           fmt.Sprintf("Your email '%s' has been verified!", email),
		"session_token": signupSessionJwt,
	}

	return respData, nil
}

func RegisterUser(sessionId, email, username, password, geolocation string) (any, error) {
	accExists, err := appModel.AccountExists(username)
	if err != nil {
		return nil, err
	}

	if accExists {
		return nil, fmt.Errorf("username error: username '%s' is unavailable", username)
	}

	hashedPassword := securityServices.HashPassword(password)

	newUser, err := user.New(email, username, hashedPassword, geolocation)
	if err != nil {
		return nil, err
	}

	authJwt := securityServices.JwtSign(appTypes.ClientUser{
		Id:       newUser.Id,
		Username: newUser.Username,
	}, os.Getenv("AUTH_JWT_SECRET"), time.Now().UTC().Add(365*24*time.Hour)) // 1 year

	go appModel.EndSignupSession(sessionId)

	respData := map[string]any{
		"msg":     "Signup success!",
		"user":    newUser,
		"authJwt": authJwt,
	}

	return respData, nil
}
