package signupControllers

import (
	"i9chat/src/appTypes"
	"i9chat/src/helpers"
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type requestNewAccountBody struct {
	Email string `json:"email"`
}

func (b requestNewAccountBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.Email,
			validation.Required,
			is.EmailFormat,
		),
	)

	return helpers.ValidationError(err, "signupControllers_validation.go", "requestNewAccountBody")
}

type verifyEmailBody struct {
	Code string `json:"code"`
}

func (b verifyEmailBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.Code,
			validation.Required,
			validation.Length(6, 6).Error("invalid code value"),
		),
	)

	return helpers.ValidationError(err, "signupControllers_validation.go", "verifyEmailBody")
}

type registerUserBody struct {
	Username    string                   `json:"username"`
	Phone       string                   `json:"phone"`
	Password    string                   `json:"password"`
	Geolocation appTypes.UserGeolocation `json:"geolocation"`
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

	return helpers.ValidationError(err, "signupControllers_validation.go", "registerUserBody")
}
