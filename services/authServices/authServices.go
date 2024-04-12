package authservices

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"model/appmodel"
	"model/usermodel"
	"services/appservices"
	"utils/helpers"

	"golang.org/x/crypto/bcrypt"
)

func RequestNewAccount(email string) (string, error) {
	// check if email already exists. if yes, send error
	accExists, err := appmodel.AccountExists(email)
	if err != nil {
		return "", err
	}

	if accExists {
		return "", fmt.Errorf("email error: an account with '%s' already exists", email)
	}

	// generate 6-digit verification code and expiration
	verfCode := rand.Intn(899999) + 100000
	expires := time.Now().UTC().Add(1 * time.Hour)

	// send 6-digit code to email
	go appservices.SendMail(email, "Email Verification", fmt.Sprintf("Your email verification code is: <b>%d</b>", verfCode))

	// store the email(varchar), verfCode(int), and vefified(bool) in an ongoing_signup table
	// get back the id as session_id
	sessionId, err := appmodel.NewSignupSession(email, verfCode)
	if err != nil {
		return "", err
	}

	// generate a 30min. JWT token that holds the session_id
	jwtToken := helpers.JwtSign(map[string]any{"sessionId": sessionId, "email": email}, os.Getenv("SIGNUP_SESSION_JWT_SECRET"), expires)

	// return the jwtToken
	return jwtToken, nil
}

func VerifyEmail(sessionId string, inputVerfCode int, email string) error {
	isSuccess, err := appmodel.VerifyEmail(sessionId, inputVerfCode)
	if err != nil {
		return err
	}

	if !isSuccess {
		return fmt.Errorf("email verification error: incorrect verification code")
	}

	go appservices.SendMail(email, "Email Verification Success", fmt.Sprintf("Your email %s has been verified!", email))

	return nil
}

func RegisterUser(sessionId string, email string, username string, password string, geolocation string) (map[string]any, string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Println(fmt.Errorf("auth service: register user: %s", err))
		return nil, "", helpers.ErrInternalServerError
	}

	accExists, err := appmodel.AccountExists(username)
	if err != nil {
		return nil, "", err
	}

	if accExists {
		return nil, "", fmt.Errorf("username error: username '%s' is unavailable", username)
	}

	userData, err := usermodel.NewUser(email, username, string(hashedPassword), geolocation)
	if err != nil {
		return nil, "", err
	}

	var user struct {
		Id       int
		Username string
	}

	helpers.ParseToStruct(userData, &user)

	// sign a jwt token
	jwtToken := helpers.JwtSign(map[string]any{
		"userId":   user.Id,
		"username": user.Username,
	}, os.Getenv("AUTH_JWT_SECRET"), time.Now().UTC().Add(365*24*time.Hour)) // 1 year

	appmodel.EndSignupSession(sessionId)

	return userData, jwtToken, nil
}

func Signin(emailOrUsername string, password string) (map[string]any, string, error) {
	userData, err := usermodel.GetUser(emailOrUsername)
	if err != nil {
		return nil, "", err
	}

	if userData == nil {
		return nil, "", fmt.Errorf("signin error: incorrect email/username or password")
	}

	hashedPassword, err := helpers.QueryRowField[string]("SELECT password FROM get_user_password($1)", emailOrUsername)
	if err != nil {
		log.Println(fmt.Errorf("authServices.go: Signin: hashed Password: get_user_password db error: %s", err))
		return nil, "", helpers.ErrInternalServerError
	}

	pwd_err := bcrypt.CompareHashAndPassword([]byte(*hashedPassword), []byte(password))
	if pwd_err != nil {
		if errors.Is(pwd_err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, "", fmt.Errorf("signin error: incorrect email/username or password")
		} else {
			log.Println(fmt.Errorf("authServices.go: Signin: %s", pwd_err))
			return nil, "", helpers.ErrInternalServerError
		}
	}

	var user struct {
		Id       int
		Username string
	}

	helpers.ParseToStruct(userData, &user)

	jwtToken := helpers.JwtSign(map[string]any{
		"userId":   user.Id,
		"username": user.Username,
	}, os.Getenv("AUTH_JWT_SECRET"), time.Now().UTC().Add(365*24*time.Hour))

	return userData, jwtToken, nil
}
