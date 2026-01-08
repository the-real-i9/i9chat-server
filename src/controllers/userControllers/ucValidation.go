package userControllers

import (
	"context"
	"errors"
	"fmt"
	"i9chat/src/appGlobals"
	"i9chat/src/appTypes"
	"i9chat/src/helpers"
	"os"
	"regexp"

	"cloud.google.com/go/storage"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/gofiber/fiber/v2"
)

type authorizePPicUploadBody struct {
	PicMIME string   `json:"pic_mime"`
	PicSize [3]int64 `json:"pic_size"` // {small, medium, large}
}

func (b authorizePPicUploadBody) Validate() error {

	err := validation.ValidateStruct(&b,
		validation.Field(&b.PicMIME, validation.Required,
			validation.In("image/jpeg", "image/png", "image/webp", "image/avif").Error(`unsupported pic_mime; use one of ["image/jpeg", "image/png", "image/webp", "image/avif"]`),
		),
		validation.Field(&b.PicSize,
			validation.Required,
			validation.Length(3, 3).Error("expected an array of 3 items"),
			validation.By(func(value any) error {
				pic_size := value.([3]int64)

				const (
					_         = iota
					SMALL int = iota - 1
					MEDIUM
					LARGE
				)

				if pic_size[SMALL] < 1*1024 || pic_size[SMALL] > 500*1024 {
					return errors.New("small pic_size out of range; min: 1KiB; max: 500KiB")
				}

				if pic_size[MEDIUM] < 1*1024 || pic_size[MEDIUM] > 1*1024*1024 {
					return errors.New("medium pic_size out of range; min: 1KiB; max: 1MeB")
				}

				if pic_size[LARGE] < 1*1024 || pic_size[LARGE] > 2*1024*1024 {
					return errors.New("large pic_size out of range; min: 1KiB; max: 2MeB")
				}

				return nil
			}),
		),
	)

	return helpers.ValidationError(err, "ucValidation.go", "authorizePPicUploadBody")
}

type changeProfilePictureBody struct {
	ProfilePicCloudName string `json:"profile_pic_cloud_name"`
}

func (b changeProfilePictureBody) Validate(ctx context.Context) error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.ProfilePicCloudName, validation.Required, validation.Match(regexp.MustCompile(
			`^small:uploads/user/profile_pics/[\w-/]+\w medium:uploads/user/profile_pics/[\w-/]+\w large:uploads/user/profile_pics/[\w-/]+\w$`,
		)).Error("invalid profile pic cloud name")),
	)

	if err != nil {
		return helpers.ValidationError(err, "userControllers_validation.go", "changeProfilePictureBody")
	}

	_, err = appGlobals.GCSClient.Bucket(os.Getenv("GCS_BUCKET_NAME")).Object(b.ProfilePicCloudName).Attrs(ctx)
	if errors.Is(err, storage.ErrObjectNotExist) {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("upload error: profile picture (%s) does not exist in cloud", b.ProfilePicCloudName))
	}

	return nil
}

type changeBioBody struct {
	NewBio string `json:"newBio"`
}

func (b changeBioBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.NewBio,
			validation.Required,
			validation.Length(1, 150).Error("maximum bio length is 150 characters"),
		),
	)

	return helpers.ValidationError(err, "userControllers_validation.go", "changeBioBody")
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
