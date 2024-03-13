package appmodel

import (
	"time"
	"utils/helpers"
)

func AccountExists(emailOrUsername string) (bool, error) {
	exist, err := helpers.QueryRowField[bool]("SELECT exist FROM account_exists($1)", emailOrUsername)
	if err != nil {
		return false, err
	}

	return *exist, nil
}

func NewSignupSession(email string, verfCode int, expires time.Time) (string, error) {
	sessionId, err := helpers.QueryRowField[string]("SELECT session_id FROM new_signup_session($1, $2, $3)", email, verfCode, expires)
	if err != nil {
		return "", err
	}

	return *sessionId, nil
}
