package userControllers

import (
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type changeProfilePictureBody struct {
	PictureData []byte `json:"pictureData"`
}

func (b changeProfilePictureBody) Validate() error {
	return validation.ValidateStruct(&b,
		validation.Field(&b.PictureData,
			validation.Required,
			validation.Length(1, 2*1024*1024).Error("maximum picture size of 2mb exceeded"),
		),
	)
}

type findNearbyUsersBody struct {
	LiveLocation string `json:"liveLocation"`
}

func (b findNearbyUsersBody) Validate() error {
	return validation.ValidateStruct(&b,
		validation.Field(&b.LiveLocation,
			validation.Required,
			validation.Match(regexp.MustCompile("^[0-9]+, ?[0-9]+, ?[0-9]+$")).Error("invalid circle format; format: pointX, pointY, radiusR"),
		),
	)
}

type updateMyGeolocationBody struct {
	NewGeolocation string `json:"newGeolocation"`
}

func (b updateMyGeolocationBody) Validate() error {
	return validation.ValidateStruct(&b,
		validation.Field(&b.NewGeolocation,
			validation.Required,
			validation.Match(regexp.MustCompile("^[0-9]+, ?[0-9]+, ?[0-9]+$")).Error("invalid circle format; format: pointX, pointY, radiusR"),
		),
	)
}
