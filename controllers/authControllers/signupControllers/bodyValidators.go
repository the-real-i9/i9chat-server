package signupControllers

import (
	"fmt"
	"i9chat/appTypes"
	"i9chat/helpers"
	"log"
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/gofiber/fiber/v2"
)

type requestNewAccountBody struct {
	Email string `json:"email"`
}

func (b requestNewAccountBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.Email,
			validation.Required,
			is.Email,
		),
	)

	if err != nil {
		if e, ok := err.(validation.InternalError); ok {
			log.Println("signup_bodyValidators.go: requestNewAccountBody", e.InternalError())
			return fiber.ErrInternalServerError
		}

		return fiber.NewError(fiber.StatusBadRequest, "validation error:", err.Error())
	}

	return nil
}

type verifyEmailBody struct {
	Code int `json:"code"`
}

func (b verifyEmailBody) Validate() error {
	mb := struct {
		Code string `json:"code"`
	}{Code: fmt.Sprint(b.Code)}

	err := validation.ValidateStruct(&mb,
		validation.Field(&mb.Code,
			validation.Required,
			validation.Length(6, 6).Error("invalid code value"),
		),
	)

	return helpers.ValidationError(err, "signup_bodyValidators.go", "verifyEmailBody")
}

type registerUserBody struct {
	Username    string                    `json:"username"`
	Password    string                    `json:"password"`
	Geolocation *appTypes.UserGeolocation `json:"geolocation"`
}

func (b registerUserBody) Validate() error {

	err := validation.ValidateStruct(&b,
		validation.Field(&b.Username,
			validation.Required,
			validation.Length(3, 0).Error("username too short"),
			validation.Match(regexp.MustCompile("^[[:alnum:]][[:alnum:]_-]+[[:alnum:]]$")).Error("invalid username syntax"),
		),
		validation.Field(&b.Password,
			validation.Required,
			validation.Length(8, 0).Error("minimum of 8 characters"),
		),
		validation.Field(&b.Geolocation, validation.Required),
	)

	return helpers.ValidationError(err, "signup_bodyValidators.go", "registerUserBody")
}
