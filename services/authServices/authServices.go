package authServices

import (
	"errors"
	"fmt"
	"i9chat/appGlobals"
	"i9chat/appTypes"
	"i9chat/helpers"
	"i9chat/models/appModel"
	user "i9chat/models/userModel"
	"i9chat/services/appServices"
	"log"
	"math/rand"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func RequestNewAccount(email string) (string, error) {
	accExists, err := appModel.AccountExists(email)
	if err != nil {
		return "", err
	}

	if accExists {
		return "", fmt.Errorf("signup error: an account with '%s' already exists", email)
	}

	verfCode, expires := rand.Intn(899999)+100000, time.Now().UTC().Add(1*time.Hour)

	go appServices.SendMail(email, "Email Verification", fmt.Sprintf("Your email verification code is: <b>%d</b>", verfCode))

	sessionId, err := appModel.NewSignupSession(email, verfCode)
	if err != nil {
		return "", err
	}

	signupSessionJwt := helpers.JwtSign(appTypes.SignupSessionData{
		SessionId: sessionId,
		Email:     email,
		Step:      "verify email",
	}, os.Getenv("SIGNUP_SESSION_JWT_SECRET"), expires)

	return signupSessionJwt, nil
}

func VerifyEmail(sessionId string, inputVerfCode int, email string) (string, error) {
	isSuccess, err := appModel.VerifyEmail(sessionId, inputVerfCode)
	if err != nil {
		return "", err
	}

	if !isSuccess {
		return "", fmt.Errorf("email verification error: incorrect verification code")
	}

	go appServices.SendMail(email, "Email Verification Success", fmt.Sprintf("Your email %s has been verified!", email))

	signupSessionJwt := helpers.JwtSign(appTypes.SignupSessionData{
		SessionId: sessionId,
		Email:     email,
		Step:      "register user",
	}, os.Getenv("SIGNUP_SESSION_JWT_SECRET"), time.Now().UTC().Add(1*time.Hour))

	return signupSessionJwt, nil
}

func RegisterUser(sessionId, email, username, password, geolocation string) (*user.User, string, error) {
	accExists, err := appModel.AccountExists(username)
	if err != nil {
		return nil, "", err
	}

	if accExists {
		return nil, "", fmt.Errorf("username error: username '%s' is unavailable", username)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Println(fmt.Errorf("authServices.go: RegisterUser: %s", err))
		return nil, "", appGlobals.ErrInternalServerError
	}

	newUser, err := user.New(email, username, string(hashedPassword), geolocation)
	if err != nil {
		return nil, "", err
	}

	authJwt := helpers.JwtSign(appTypes.ClientUser{
		Id:       newUser.Id,
		Username: newUser.Username,
	}, os.Getenv("AUTH_JWT_SECRET"), time.Now().UTC().Add(365*24*time.Hour)) // 1 year

	go appModel.EndSignupSession(sessionId)

	return newUser, authJwt, nil
}

func Signin(emailOrUsername string, password string) (*user.User, string, error) {
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

	cmp_err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if cmp_err != nil {
		if errors.Is(cmp_err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, "", fmt.Errorf("signin error: incorrect email/username or password")
		} else {
			log.Println(fmt.Errorf("authServices.go: Signin: %s", cmp_err))
			return nil, "", appGlobals.ErrInternalServerError
		}
	}

	authJwt := helpers.JwtSign(appTypes.ClientUser{
		Id:       theUser.Id,
		Username: theUser.Username,
	}, os.Getenv("AUTH_JWT_SECRET"), time.Now().UTC().Add(365*24*time.Hour))

	return theUser, authJwt, nil
}
