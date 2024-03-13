package authservices

import (
	"fmt"
	"math/rand"
	appmodel "model/app"
	"os"
	appservices "services/app"
	"time"
	"utils/helpers"
)

func RequestNewAccount(email string) (string, error) {
	// check if email already exists. if yes, send error
	accExists, err := appmodel.AccountExists(email)
	if err != nil {
		return "", err
	}

	if accExists {
		return "", fmt.Errorf(`an account with "%s" already exists`, email)
	}

	// generate 6-digit verification code and expiration
	verfCode := rand.Intn(899999) + 100000
	expires := time.Now().Add(30 * time.Minute)

	// send 6-digit code to email
	mail_err := appservices.SendMail(email, "i9chat - Email Verification", fmt.Sprintf("Your email verification code is: <b>%d</b>", verfCode))
	if mail_err != nil {
		return "", mail_err
	}

	// store the email(varchar), verfCode(int), and vefified(bool) in an ongoing_signup table
	// get back the id as session_id
	sessionId, err := appmodel.NewSignupSession(email, verfCode, expires)
	if err != nil {
		return "", nil
	}

	// generate a 30min. JWT token that holds the session_id
	jwtToken := helpers.JwtSign(map[string]any{"sessionId": sessionId}, os.Getenv("SIGNUP_JWT_SECRET"), expires)

	// return the jwtToken
	return jwtToken, nil
}
