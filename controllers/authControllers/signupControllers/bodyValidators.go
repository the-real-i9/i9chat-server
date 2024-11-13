package signupControllers

import (
	"fmt"
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type requestNewAccountBody struct {
	Email string `json:"email"`
}

func (b requestNewAccountBody) Validate() error {
	return validation.ValidateStruct(&b,
		validation.Field(&b.Email,
			validation.Required,
			is.Email,
		),
	)
}

type verifyEmailBody struct {
	Code int `json:"code"`
}

func (b verifyEmailBody) Validate() error {
	mb := struct {
		Code string `json:"code"`
	}{Code: fmt.Sprint(b.Code)}

	return validation.ValidateStruct(&mb,
		validation.Field(&mb.Code,
			validation.Required,
			validation.Length(6, 6).Error("invalid code value"),
		),
	)
}

type registerUserBody struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	Geolocation string `json:"geolocation"`
}

func (b registerUserBody) Validate() error {

	return validation.ValidateStruct(&b,
		validation.Field(&b.Username,
			validation.Required,
			validation.Length(3, 0).Error("username too short"),
			validation.Match(regexp.MustCompile("^[[:alnum:]][[:alnum:]_-]+[[:alnum:]]$")).Error("invalid username syntax"),
		),
		validation.Field(&b.Password,
			validation.Required,
			validation.Length(8, 0).Error("minimum of 8 characters"),
		),
		validation.Field(&b.Geolocation,
			validation.Required,
			validation.Match(regexp.MustCompile("^[0-9]+, ?[0-9]+, ?[0-9]+$")).Error("invalid circle format; format: pointX, pointY, radiusR"),
		),
	)
}
