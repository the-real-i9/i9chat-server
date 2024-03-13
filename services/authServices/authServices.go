package authservices

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"model/appmodel"
	"services/appservices"
	"utils/helpers"
)

func RequestNewAccount(email string) (string, error) {
	// check if email already exists. if yes, send error
	accExists, err := appmodel.AccountExists(email)
	if err != nil {
		log.Println(err)
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
	sessionId, err := appmodel.NewSignupSession(email, verfCode, expires)
	if err != nil {
		return "", fmt.Errorf("new signup session error: %s", err)
	}

	// generate a 30min. JWT token that holds the session_id
	jwtToken := helpers.JwtSign(map[string]any{"sessionId": sessionId}, os.Getenv("SIGNUP_JWT_SECRET"), expires)

	// return the jwtToken
	return jwtToken, nil
}
