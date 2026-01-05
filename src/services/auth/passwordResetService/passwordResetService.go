package passwordResetService

import (
	"context"
	"fmt"
	"i9chat/src/appErrors/userErrors"
	"i9chat/src/helpers"
	user "i9chat/src/models/userModel"
	"i9chat/src/services/mailService"
	"i9chat/src/services/securityServices"
	"time"

	"github.com/gofiber/fiber/v2"
)

type passReset1RespT struct {
	Msg string `json:"msg"`
}

func RequestPasswordReset(ctx context.Context, email string) (passReset1RespT, map[string]any, error) {
	var resp passReset1RespT

	userExists, err := user.Exists(ctx, email)
	if err != nil {
		return resp, nil, err
	}

	if !userExists {
		return resp, nil, fiber.NewError(fiber.StatusNotFound, userErrors.NonExistingUser)
	}

	pwdrToken, expires := securityServices.GenerateTokenCodeExp()

	go mailService.SendMail(email, "Confirm your email: Password Reset", fmt.Sprintf("<p>Your password reset token is <strong>%s</strong>.</p>", pwdrToken))

	sessionData := map[string]any{
		"email":            email,
		"pwdrToken":        pwdrToken,
		"pwdrTokenExpires": expires,
	}

	resp.Msg = fmt.Sprintf("Enter the 6-digit number token sent to %s to reset your password", email)

	return resp, sessionData, nil
}

type passReset2RespT struct {
	Msg string `json:"msg"`
}

func ConfirmEmail(ctx context.Context, sessionData map[string]any, inputResetToken string) (passReset2RespT, map[string]any, error) {
	var resp passReset2RespT

	sd := helpers.ToStruct[struct {
		Email            string
		PwdrToken        string
		PwdrTokenExpires time.Time
	}](sessionData)

	if sd.PwdrToken != inputResetToken {
		return resp, nil, fiber.NewError(fiber.StatusBadRequest, userErrors.IncorrectResetToken)
	}

	if sd.PwdrTokenExpires.Before(time.Now()) {
		return resp, nil, fiber.NewError(fiber.StatusBadRequest, userErrors.ResetTokenExpired)
	}

	newSessionData := map[string]any{"email": sd.Email}

	resp.Msg = fmt.Sprintf("%s, you're about to reset your password!", sd.Email)

	return resp, newSessionData, nil
}

type passReset3RespT struct {
	Msg string `json:"msg"`
}

func ResetPassword(ctx context.Context, sessionData map[string]any, newPassword string) (passReset3RespT, error) {
	var resp passReset3RespT

	email := sessionData["email"].(string)

	hashedPassword, err := securityServices.HashPassword(newPassword)
	if err != nil {
		return resp, err
	}

	done, err := user.ChangePassword(ctx, email, hashedPassword)
	if err != nil {
		return resp, err
	}

	if done {
		go mailService.SendMail(email, "Password Reset Success", fmt.Sprintf("<p>%s, your password has been changed successfully!</p>", email))

		resp.Msg = "Your password has been changed successfully"
	}

	return resp, nil
}
