package passwordResetService

import (
	"context"
	"fmt"
	"i9chat/src/helpers"
	user "i9chat/src/models/userModel"
	"i9chat/src/services/mailService"
	"i9chat/src/services/securityServices"
	"time"

	"github.com/gofiber/fiber/v2"
)

func RequestPasswordReset(ctx context.Context, email string) (any, map[string]any, error) {

	userExists, err := user.Exists(ctx, email)
	if err != nil {
		return nil, nil, err
	}

	if !userExists {
		return nil, nil, fiber.NewError(fiber.StatusNotFound, "No user with this email exists.")
	}

	pwdrToken, expires := securityServices.GenerateTokenCodeExp()

	go mailService.SendMail(email, "Confirm your email: Password Reset", fmt.Sprintf("<p>Your password reset token is <strong>%s</strong>.</p>", pwdrToken))

	sessionData := map[string]any{
		"email":            email,
		"pwdrToken":        pwdrToken,
		"pwdrTokenExpires": expires,
	}

	respData := map[string]any{
		"msg": fmt.Sprintf("Enter the 6-digit number token sent to %s to reset your password", email),
	}

	return respData, sessionData, nil
}

func ConfirmEmail(ctx context.Context, sessionData map[string]any, inputResetToken string) (any, map[string]any, error) {
	var sd struct {
		Email            string
		PwdrToken        string
		PwdrTokenExpires time.Time
	}

	helpers.ToStruct(sessionData, &sd)

	if sd.PwdrToken != inputResetToken {
		return "", nil, fiber.NewError(fiber.StatusBadRequest, "Incorrect password reset token! Check or Re-submit your email.")
	}

	if sd.PwdrTokenExpires.Before(time.Now()) {
		return "", nil, fiber.NewError(fiber.StatusBadRequest, "Password reset token expired! Re-submit your email.")
	}

	newSessionData := map[string]any{"email": sd.Email}

	respData := map[string]any{
		"msg": fmt.Sprintf("%s, you're about to reset your password!", sd.Email),
	}

	return respData, newSessionData, nil
}

func ResetPassword(ctx context.Context, sessionData map[string]any, newPassword string) (map[string]any, error) {
	email := sessionData["email"].(string)

	hashedPassword, err := securityServices.HashPassword(newPassword)
	if err != nil {
		return nil, err
	}

	m_err := user.ChangePassword(ctx, email, hashedPassword)
	if m_err != nil {
		return nil, m_err
	}

	go mailService.SendMail(email, "Password Reset Success", fmt.Sprintf("<p>%s, your password has been changed successfully!</p>", email))

	respData := map[string]any{
		"msg": "Your password has been changed successfully",
	}

	return respData, nil
}
