package userControllers

import (
	"i9chat/src/appTypes"
	"i9chat/src/helpers"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type changeProfilePictureBody struct {
	PictureData []byte `json:"pictureData"`
}

func (b changeProfilePictureBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.PictureData,
			validation.Required,
			validation.Length(1, 2*1024*1024).Error("maximum picture size of 2mb exceeded"),
		),
	)

	return helpers.ValidationError(err, "userControllers_validation.go", "changeProfilePictureBody")
}

type changePhoneBody struct {
	Phone string `json:"newPhoneNumber"`
}

func (b changePhoneBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.Phone, validation.Required),
	)

	return helpers.ValidationError(err, "userControllers_validation.go", "changePhoneBody")

}

type findNearbyUsersQuery struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Radius float64 `json:"radius"`
}

func (b findNearbyUsersQuery) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.X, validation.Required),
		validation.Field(&b.Y, validation.Required),
		validation.Field(&b.Radius, validation.Required),
	)

	return helpers.ValidationError(err, "userControllers_validation.go", "findNearbyUsersQuery")

}

type updateMyGeolocationBody struct {
	NewGeolocation appTypes.UserGeolocation `json:"newGeolocation"`
}

func (b updateMyGeolocationBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.NewGeolocation, validation.Required),
	)

	return helpers.ValidationError(err, "userControllers_validation.go", "updateMyGeolocationBody")

}

type searchUserQuery struct {
	EmailUsernamePhone string `json:"emailUsernamePhone"`
}

func (b searchUserQuery) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.EmailUsernamePhone, validation.Required),
	)

	return helpers.ValidationError(err, "userControllers_validation.go", "searchUserQuery")

}
