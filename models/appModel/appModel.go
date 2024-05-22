package appModel

import (
	"fmt"
	"i9chat/utils/helpers"
	"log"
)

func AccountExists(emailOrUsername string) (bool, error) {
	exist, err := helpers.QueryRowField[bool]("SELECT exist FROM account_exists($1)", emailOrUsername)

	if err != nil {
		log.Println(fmt.Errorf("appModel.go: AccountExists: %s", err))
		return false, helpers.ErrInternalServerError
	}

	return *exist, nil
}

func NewSignupSession(email string, verfCode int) (string, error) {
	sessionId, err := helpers.QueryRowField[string]("SELECT session_id FROM new_signup_session($1, $2)", email, verfCode)

	if err != nil {
		log.Println(fmt.Errorf("appModel.go: NewSignupSession: %s", err))
		return "", helpers.ErrInternalServerError
	}

	return *sessionId, nil
}

func VerifyEmail(sessionId string, verfCode int) (bool, error) {
	isSuccess, err := helpers.QueryRowField[bool]("SELECT is_success FROM verify_email($1, $2)", sessionId, verfCode)

	if err != nil {
		log.Println(fmt.Errorf("appModel.go: VerifyEmail: %s", err))
		return false, helpers.ErrInternalServerError
	}

	return *isSuccess, nil
}

func EndSignupSession(sessionId string) {
	go helpers.QueryRowField[bool]("SELECT end_signup_session ($1)", sessionId)
}
