package appmodel

import (
	"utils/helpers"
)

func AccountExists(emailOrUsername string) (bool, error) {
	exist, err := helpers.QueryRowField[bool]("SELECT exist FROM account_exists($1)", emailOrUsername)
	if err != nil {
		return false, err
	}

	return *exist, nil
}

func NewSignupSession(email string, verfCode int) (string, error) {
	sessionId, err := helpers.QueryRowField[string]("SELECT session_id FROM new_signup_session($1, $2)", email, verfCode)
	if err != nil {
		return "", err
	}

	return *sessionId, nil
}

func VerifyEmail(sessionId string, verfCode int) (bool, error) {
	isSuccess, err := helpers.QueryRowField[bool]("SELECT is_success FROM verify_email($1, $2)", sessionId, verfCode)
	if err != nil {
		return false, err
	}

	return *isSuccess, nil
}
