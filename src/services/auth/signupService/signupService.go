package signupService

import (
	"context"
	"fmt"
	"i9chat/src/appErrors/userErrors"
	"i9chat/src/appTypes"
	"i9chat/src/appTypes/UITypes"
	"i9chat/src/helpers"
	"i9chat/src/services/cloudStorageService"
	"i9chat/src/services/mailService"
	"i9chat/src/services/securityServices"
	"i9chat/src/services/userService"
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/vmihailenco/msgpack/v5"
)

type signup1RespT struct {
	Msg string `msgpack:"msg"`
}

func RequestNewAccount(ctx context.Context, email string) (signup1RespT, map[string]any, error) {
	var resp signup1RespT

	userExists, err := userService.UserExists(ctx, email)
	if err != nil {
		return resp, nil, err
	}

	if userExists {
		return resp, nil, fiber.NewError(fiber.StatusConflict, userErrors.EmailAlreadyExists)
	}

	verfCode, expires := securityServices.GenerateTokenCodeExp()

	go mailService.SendMail(email, "Email Verification", fmt.Sprintf("Your email verification code is: <b>%s</b>", verfCode))

	sessionData := map[string]any{
		"email":        email,
		"vCode":        verfCode,
		"vCodeExpires": expires,
	}

	resp.Msg = "A 6-digit verification code has been sent to " + email

	return resp, sessionData, nil
}

type signup2RespT struct {
	Msg string `msgpack:"msg"`
}

func VerifyEmail(ctx context.Context, sessionData msgpack.RawMessage, inputVerfCode string) (signup2RespT, map[string]any, error) {
	var resp signup2RespT

	sd := helpers.FromBtMsgPack[struct {
		Email        string    `msgpack:"email"`
		VCode        string    `msgpack:"vCode"`
		VCodeExpires time.Time `msgpack:"vCodeExpires"`
	}](sessionData)

	if sd.VCode != inputVerfCode {
		return resp, nil, fiber.NewError(fiber.StatusBadRequest, userErrors.IncorrectVerfCode)
	}

	if sd.VCodeExpires.Before(time.Now()) {
		return resp, nil, fiber.NewError(fiber.StatusBadRequest, userErrors.VerfCodeExpired)
	}

	go mailService.SendMail(sd.Email, "Email Verification Success", fmt.Sprintf("Your email %s has been verified!", sd.Email))

	newSessionData := map[string]any{"email": sd.Email}

	resp.Msg = fmt.Sprintf("Your email '%s' has been verified!", sd.Email)

	return resp, newSessionData, nil
}

type signup3RespT struct {
	Msg  string             `msgpack:"msg"`
	User UITypes.ClientUser `msgpack:"user"`
}

func RegisterUser(ctx context.Context, sessionData msgpack.RawMessage, username, password, bio string) (signup3RespT, string, error) {
	var resp signup3RespT

	email := helpers.FromBtMsgPack[struct {
		Email string `msgpack:"email"`
	}](sessionData).Email

	userExists, err := userService.UserExists(ctx, username)
	if err != nil {
		return resp, "", err
	}

	if userExists {
		return resp, "", fiber.NewError(fiber.StatusConflict, userErrors.UsernameUnavailable)
	}

	hashedPassword, err := securityServices.HashPassword(password)
	if err != nil {
		return resp, "", err
	}

	if bio == "" {
		bio = "I love i9chat!"
	}

	newUser, err := userService.NewUser(ctx, email, username, hashedPassword, bio)
	if err != nil {
		return resp, "", err
	}

	authJwt, err := securityServices.JwtSign(appTypes.ClientUser{
		Username: newUser.Username,
	}, os.Getenv("AUTH_JWT_SECRET"), time.Now().UTC().Add(10*24*time.Hour)) // 10 days
	if err != nil {
		return resp, "", err
	}

	newUser.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(newUser.ProfilePicUrl)

	resp.Msg = "Signup success!"
	resp.User = UITypes.ClientUser{Username: newUser.Username, ProfilePicUrl: newUser.ProfilePicUrl, Presence: newUser.Presence}

	return resp, authJwt, nil
}
