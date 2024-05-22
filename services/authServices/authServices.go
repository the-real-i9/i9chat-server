package authServices

import (
	"errors"
	"fmt"
	"i9chat/models/appModel"
	"i9chat/models/userModel"
	"i9chat/services/appServices"
	"i9chat/utils/helpers"
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

	signupSessionJwt := helpers.JwtSign(map[string]any{
		"sessionId": sessionId,
		"email":     email,
	}, os.Getenv("SIGNUP_SESSION_JWT_SECRET"), expires)

	return signupSessionJwt, nil
}

func VerifyEmail(sessionId string, inputVerfCode int, email string) error {
	isSuccess, err := appModel.VerifyEmail(sessionId, inputVerfCode)
	if err != nil {
		return err
	}

	if !isSuccess {
		return fmt.Errorf("email verification error: incorrect verification code")
	}

	go appServices.SendMail(email, "Email Verification Success", fmt.Sprintf("Your email %s has been verified!", email))

	return nil
}

func RegisterUser(sessionId string, email string, username string, password string, geolocation string) (map[string]any, string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Println(fmt.Errorf("authServices.go: RegisterUser: %s", err))
		return nil, "", helpers.ErrInternalServerError
	}

	accExists, err := appModel.AccountExists(username)
	if err != nil {
		return nil, "", err
	}

	if accExists {
		return nil, "", fmt.Errorf("username error: username '%s' is unavailable", username)
	}

	userData, err := userModel.NewUser(email, username, string(hashedPassword), geolocation)
	if err != nil {
		return nil, "", err
	}

	var user struct {
		Id       int
		Username string
	}

	helpers.ParseToStruct(userData, &user)

	authJwt := helpers.JwtSign(map[string]any{
		"userId":   user.Id,
		"username": user.Username,
	}, os.Getenv("AUTH_JWT_SECRET"), time.Now().UTC().Add(365*24*time.Hour)) // 1 year

	appModel.EndSignupSession(sessionId)

	return userData, authJwt, nil
}

func Signin(emailOrUsername string, password string) (map[string]any, string, error) {
	userData, err := userModel.GetUser(emailOrUsername)
	if err != nil {
		return nil, "", err
	}

	if userData == nil {
		return nil, "", fmt.Errorf("signin error: incorrect email/username or password")
	}

	hashedPassword, err := helpers.QueryRowField[string]("SELECT password FROM get_user_password($1)", emailOrUsername)
	if err != nil {
		log.Println(fmt.Errorf("authServices.go: Signin: DB query error: get_user_password(): %s", err))
		return nil, "", helpers.ErrInternalServerError
	}

	cmp_err := bcrypt.CompareHashAndPassword([]byte(*hashedPassword), []byte(password))
	if cmp_err != nil {
		if errors.Is(cmp_err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, "", fmt.Errorf("signin error: incorrect email/username or password")
		} else {
			log.Println(fmt.Errorf("authServices.go: Signin: %s", cmp_err))
			return nil, "", helpers.ErrInternalServerError
		}
	}

	var user struct {
		Id       int
		Username string
	}

	helpers.ParseToStruct(userData, &user)

	authJwt := helpers.JwtSign(map[string]any{
		"userId":   user.Id,
		"username": user.Username,
	}, os.Getenv("AUTH_JWT_SECRET"), time.Now().UTC().Add(365*24*time.Hour))

	return userData, authJwt, nil
}
