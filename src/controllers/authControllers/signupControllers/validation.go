package signupControllers

import (
	"i9chat/src/helpers"
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type requestNewAccountBody struct {
	Email string `msgpack:"email"`
}

func (b requestNewAccountBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.Email,
			validation.Required,
			is.EmailFormat.Error("incorrect email format"),
		),
	)

	return helpers.ValidationError(err, "signupControllers_validation.go", "requestNewAccountBody")
}

type verifyEmailBody struct {
	Code string `msgpack:"code"`
}

func (b verifyEmailBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.Code, validation.Required),
	)

	return helpers.ValidationError(err, "signupControllers_validation.go", "verifyEmailBody")
}

type registerUserBody struct {
	Username string `msgpack:"username"`
	Password string `msgpack:"password"`
	Bio      string `msgpack:"bio"`
}

func (b registerUserBody) Validate() error {

	err := validation.ValidateStruct(&b,
		validation.Field(&b.Username,
			validation.Required,
			validation.Length(3, 0).Error("username too short"),
			validation.Match(regexp.MustCompile("^[[:alnum:]][[:alnum:]_-]+[[:alnum:]]$")).Error("username contains invalid characters"),
		),
		validation.Field(&b.Password,
			validation.Required,
			validation.Length(8, 0).Error("password too short. minimum of 8 characters"),
		),
	)

	return helpers.ValidationError(err, "signupControllers_validation.go", "registerUserBody")
}
