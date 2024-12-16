package userControllers

import (
	"i9chat/helpers"
	"regexp"

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

	return helpers.ValidationError(err, "user_bodyValidators.go", "changeProfilePictureBody")

}

type findNearbyUsersBody struct {
	LiveLocation string `json:"liveLocation"`
}

func (b findNearbyUsersBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.LiveLocation,
			validation.Required,
			validation.Match(regexp.MustCompile("^[0-9]+, ?[0-9]+, ?[0-9]+$")).Error("invalid circle format; format: pointX, pointY, radiusR"),
		),
	)

	return helpers.ValidationError(err, "user_bodyValidators.go", "findNearbyUsersBody")

}

type updateMyGeolocationBody struct {
	NewGeolocation string `json:"newGeolocation"`
}

func (b updateMyGeolocationBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.NewGeolocation,
			validation.Required,
			validation.Match(regexp.MustCompile("^[0-9]+, ?[0-9]+, ?[0-9]+$")).Error("invalid circle format; format: pointX, pointY, radiusR"),
		),
	)

	return helpers.ValidationError(err, "user_bodyValidators.go", "updateMyGeolocationBody")

}
