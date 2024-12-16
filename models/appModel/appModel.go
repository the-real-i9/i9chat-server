package appModel

import (
	"context"
	"i9chat/appGlobals"
	"i9chat/helpers"
	"log"
)

func AccountExists(ctx context.Context, emailOrUsername string) (bool, error) {
	exist, err := helpers.QueryRowField[bool](ctx, "SELECT exist FROM account_exists($1)", emailOrUsername)

	if err != nil {
		log.Println("appModel.go: AccountExists:", err)
		return false, appGlobals.ErrInternalServerError
	}

	return *exist, nil
}

func NewSignupSession(ctx context.Context, email string, verfCode int) (string, error) {
	sessionId, err := helpers.QueryRowField[string](ctx, "SELECT session_id FROM new_signup_session($1, $2)", email, verfCode)

	if err != nil {
		log.Println("appModel.go: NewSignupSession:", err)
		return "", appGlobals.ErrInternalServerError
	}

	return *sessionId, nil
}

func VerifyEmail(ctx context.Context, sessionId string, verfCode int) (bool, error) {
	isSuccess, err := helpers.QueryRowField[bool](ctx, "SELECT is_success FROM verify_email($1, $2)", sessionId, verfCode)

	if err != nil {
		log.Println("appModel.go: VerifyEmail:", err)
		return false, appGlobals.ErrInternalServerError
	}

	return *isSuccess, nil
}

func EndSignupSession(ctx context.Context, sessionId string) {
	helpers.QueryRowField[bool](ctx, "SELECT end_signup_session ($1)", sessionId)
}
