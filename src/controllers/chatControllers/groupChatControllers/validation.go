package groupChatControllers

import (
	"context"
	"errors"
	"fmt"
	"i9chat/src/appGlobals"
	"i9chat/src/helpers"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/gofiber/fiber/v2"
)

type newGroupChatBody struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	PictureData []byte   `json:"pictureData"`
	InitUsers   []string `json:"initUsers"`
	CreatedAt   int64    `json:"createdAt"`
}

func (b newGroupChatBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.Name, validation.Required),
		validation.Field(&b.Description, validation.Required),
		validation.Field(&b.PictureData,
			validation.Required,
			validation.Length(1024, 2*1024*1024).Error("group picture size out of range. min: 1KiB. max: 2MiB"),
		),
		validation.Field(&b.InitUsers, validation.Required, validation.Length(1, 0).Error("at least 1 other user is required to start a group")),
		validation.Field(&b.CreatedAt, validation.Required, validation.Max(time.Now().UTC().UnixMilli()).Error("invalid future time")),
	)

	return helpers.ValidationError(err, "groupChat_validation.go", "newGroupChatBody")

}

type changeGroupNameAction struct {
	NewName string `json:"newName"`
}

func (d changeGroupNameAction) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.NewName, validation.Required),
	)

	return helpers.ValidationError(err, "groupChat_validation.go", "changeGroupNameAction")

}

type changeGroupDescriptionAction struct {
	NewDescription string `json:"newDescription"`
}

func (d changeGroupDescriptionAction) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.NewDescription, validation.Required),
	)

	return helpers.ValidationError(err, "groupChat_validation.go", "changeGroupDescriptionAction")

}

type authorizeGroupPicUploadBody struct {
	PicMIME string   `json:"pic_mime"`
	PicSize [3]int64 `json:"pic_size"` // {small, medium, large}
}

func (b authorizeGroupPicUploadBody) Validate() error {

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

				if pic_size[SMALL] < 100*1024 || pic_size[SMALL] > 500*1024 {
					return errors.New("small pic_size out of range; min: 100KiB; max: 500KiB")
				}

				if pic_size[MEDIUM] < 500*1024 || pic_size[MEDIUM] > 2*1024*1024 {
					return errors.New("medium pic_size out of range; min: 500KiB; max: 2MeB")
				}

				if pic_size[LARGE] < 2*1024*1024 || pic_size[LARGE] > 5*1024*1024 {
					return errors.New("medium pic_size out of range; min: 2MeB; max: 5MeB")
				}

				return nil
			}),
		),
	)

	return helpers.ValidationError(err, "groupChatControllers_validation.go", "authorizeGroupPicUploadBody")
}

type changeGroupPictureAction struct {
	PicCloudName string `json:"pic_cloud_name"`
}

func (d changeGroupPictureAction) Validate(ctx context.Context) error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.PicCloudName, validation.Required, validation.By(func(value any) error {
			val := value.(string)

			if !(strings.HasPrefix(val, "small:uploads/group/group_pics/") && strings.Contains(val, " medium:uploads/group/group_pics/") && strings.HasSuffix(val, " large:uploads/group/group_pics/")) {
				return errors.New("invalid group pic cloud name")
			}

			return nil
		})),
	)

	if err != nil {
		return helpers.ValidationError(err, "groupChat_validation.go", "changeGroupPictureAction")
	}

	_, err = appGlobals.GCSClient.Bucket(os.Getenv("GCS_BUCKET_NAME")).Object(d.PicCloudName).Attrs(ctx)
	if errors.Is(err, storage.ErrObjectNotExist) {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("upload error: group picture (%s) does not exist in cloud", d.PicCloudName))
	}

	return nil
}

type addUsersToGroupAction struct {
	NewUsers []string `json:"newUsers"`
}

func (d addUsersToGroupAction) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.NewUsers, validation.Required, validation.Length(1, 0).Error("no users provided")),
	)

	return helpers.ValidationError(err, "groupChat_validation.go", "addUsersToGroupAction")

}

type actOnSingleUserAction struct {
	User string `json:"user"`
}

func (d actOnSingleUserAction) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.User, validation.Required),
	)

	return helpers.ValidationError(err, "groupChat_validation.go", "actOnSingleUserAction")

}
