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

type updateMyGeolocationBody struct {
	NewGeolocation appTypes.UserGeolocation `json:"newGeolocation"`
}

func (b updateMyGeolocationBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.NewGeolocation, validation.Required),
	)

	return helpers.ValidationError(err, "userControllers_validation.go", "updateMyGeolocationBody")

}
