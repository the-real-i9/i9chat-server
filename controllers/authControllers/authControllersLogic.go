package authControllers

import (
	"fmt"
	"i9chat/appTypes"
	"i9chat/models/appModel"
	user "i9chat/models/userModel"
	"i9chat/services/authServices"
	"i9chat/services/mailService"
	"os"
	"time"
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

	go mailService.SendMail(email, "Email Verification", fmt.Sprintf("Your email verification code is: <b>%d</b>", verfCode))

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

func verifyEmail(sessionId string, inputVerfCode int, email string) (string, error) {
	isSuccess, err := appModel.VerifyEmail(sessionId, inputVerfCode)
	if err != nil {
		return "", err
	}

	if !isSuccess {
		return "", fmt.Errorf("email verification error: incorrect verification code")
	}

	go mailService.SendMail(email, "Email Verification Success", fmt.Sprintf("Your email %s has been verified!", email))

	signupSessionJwt := authServices.JwtSign(appTypes.SignupSessionData{
		SessionId: sessionId,
		Email:     email,
		Step:      "register user",
	}, os.Getenv("SIGNUP_SESSION_JWT_SECRET"), time.Now().UTC().Add(1*time.Hour))

	return signupSessionJwt, nil
}

func registerUser(sessionId, email, username, password, geolocation string) (*user.User, string, error) {
	accExists, err := appModel.AccountExists(username)
	if err != nil {
		return nil, "", err
	}

	if accExists {
		return nil, "", fmt.Errorf("username error: username '%s' is unavailable", username)
	}

	hashedPassword := authServices.HashPassword(password)

	newUser, err := user.New(email, username, hashedPassword, geolocation)
	if err != nil {
		return nil, "", err
	}

	authJwt := authServices.JwtSign(appTypes.ClientUser{
		Id:       newUser.Id,
		Username: newUser.Username,
	}, os.Getenv("AUTH_JWT_SECRET"), time.Now().UTC().Add(365*24*time.Hour)) // 1 year

	go appModel.EndSignupSession(sessionId)

	return newUser, authJwt, nil
}

func signin(emailOrUsername string, password string) (*user.User, string, error) {
	theUser, err := user.FindOne(emailOrUsername)
	if err != nil {
		return nil, "", err
	}

	if theUser == nil {
		return nil, "", fmt.Errorf("signin error: incorrect email/username or password")
	}

	hashedPassword, err := user.GetPassword(emailOrUsername)
	if err != nil {
		return nil, "", err
	}

	if !authServices.PasswordMatchesHash(hashedPassword, password) {
		return nil, "", fmt.Errorf("signin error: incorrect email/username or password")
	}

	authJwt := authServices.JwtSign(appTypes.ClientUser{
		Id:       theUser.Id,
		Username: theUser.Username,
	}, os.Getenv("AUTH_JWT_SECRET"), time.Now().UTC().Add(365*24*time.Hour))

	return theUser, authJwt, nil
}
