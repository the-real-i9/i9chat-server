package signinControllers

import (
	"i9chat/helpers"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type signInBody struct {
	EmailOrUsername string `json:"emailOrUsername"`
	Password        string `json:"password"`
}

func (b signInBody) Validate() error {

	err := validation.ValidateStruct(&b,
		validation.Field(&b.EmailOrUsername,
			validation.Required,
			validation.When(strings.ContainsAny(b.EmailOrUsername, "@"),
				is.Email.Error("invalid email or username"),
			).Else(
				validation.Length(3, 0).Error("invalid email or username"),
			),
		),
		validation.Field(&b.Password,
			validation.Required,
		),
	)

	return helpers.ValidationError(err, "signin_bodyValidators.go", "signInBody")

}
