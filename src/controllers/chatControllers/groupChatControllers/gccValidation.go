package groupChatControllers

import (
	"context"
	"errors"
	"fmt"
	"i9chat/src/controllers/chatControllers/chatTypes"
	"i9chat/src/helpers"
	"i9chat/src/services/cloudStorageService"
	"regexp"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type newGroupChatBody struct {
	Name             string   `msgpack:"name"`
	Description      string   `msgpack:"description"`
	PictureCloudName string   `msgpack:"pictureCloudName"`
	InitUsers        []string `msgpack:"initUsers"`
	CreatedAt        int64    `msgpack:"createdAt"`
}

func (b newGroupChatBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.Name, validation.Required),
		validation.Field(&b.Description, validation.Required),
		validation.Field(&b.PictureCloudName, validation.Required,
			validation.Match(regexp.MustCompile(
				`^small:uploads/group/group_pics/[\w-/]+\w medium:uploads/group/group_pics/[\w-/]+\w large:uploads/group/group_pics/[\w-/]+\w$`,
			)).Error("invalid group picture cloud name"),
		),
		validation.Field(&b.InitUsers, validation.Required, validation.Length(1, 0).Error("at least 1 other user is required to start a group")),
		validation.Field(&b.CreatedAt, validation.Required, validation.Max(time.Now().UTC().UnixMilli()).Error("invalid future time")),
	)

	if err != nil {
		return helpers.ValidationError(err, "gccValidation.go", "newGroupChatBody")
	}

	go func(gpicCn string) {
		ctx := context.Background()

		var (
			smallPPicCn  string
			mediumPPicCn string
			largePPicCn  string
		)

		fmt.Sscanf(gpicCn, "small:%s medium:%s large:%s", &smallPPicCn, &mediumPPicCn, &largePPicCn)

		if mInfo := cloudStorageService.GetMediaInfo(ctx, smallPPicCn); mInfo != nil {
			if mInfo.Size < 1*1024 || mInfo.Size > 500*1024 {
				cloudStorageService.DeleteCloudMedia(ctx, smallPPicCn)
			}
		}

		if mInfo := cloudStorageService.GetMediaInfo(ctx, mediumPPicCn); mInfo != nil {
			if mInfo.Size < 1*1024 || mInfo.Size > 1*1024*1024 {
				cloudStorageService.DeleteCloudMedia(ctx, mediumPPicCn)
			}
		}

		if mInfo := cloudStorageService.GetMediaInfo(ctx, largePPicCn); mInfo != nil {
			if mInfo.Size < 1*1024 || mInfo.Size > 2*1024*1024 {
				cloudStorageService.DeleteCloudMedia(ctx, largePPicCn)
			}
		}
	}(b.PictureCloudName)

	return nil
}

type changeGroupNameAction struct {
	NewName string `msgpack:"newName"`
}

func (d changeGroupNameAction) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.NewName, validation.Required),
	)

	return helpers.ValidationError(err, "gccValidation.go", "changeGroupNameAction")

}

type changeGroupDescriptionAction struct {
	NewDescription string `msgpack:"newDescription"`
}

func (d changeGroupDescriptionAction) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.NewDescription, validation.Required),
	)

	return helpers.ValidationError(err, "gccValidation.go", "changeGroupDescriptionAction")

}

type authorizeGroupPicUploadBody struct {
	PicMIME string   `msgpack:"pic_mime"`
	PicSize [3]int64 `msgpack:"pic_size"` // {small, medium, large}
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

	return helpers.ValidationError(err, "gccValidation.go", "authorizeGroupPicUploadBody")
}

type changeGroupPictureAction struct {
	PictureCloudName string `msgpack:"picture_cloud_name"`
}

func (d changeGroupPictureAction) Validate(ctx context.Context) error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.PictureCloudName, validation.Required, validation.Match(regexp.MustCompile(
			`^small:uploads/group/group_pics/[\w-/]+\w medium:uploads/group/group_pics/[\w-/]+\w large:uploads/group/group_pics/[\w-/]+\w$`,
		)).Error("invalid group picture cloud name")),
	)

	if err != nil {
		return helpers.ValidationError(err, "gccValidation.go", "changeGroupPictureAction")
	}

	go func(gpicCn string) {
		ctx := context.Background()

		var (
			smallPPicCn  string
			mediumPPicCn string
			largePPicCn  string
		)

		fmt.Sscanf(gpicCn, "small:%s medium:%s large:%s", &smallPPicCn, &mediumPPicCn, &largePPicCn)

		if mInfo := cloudStorageService.GetMediaInfo(ctx, smallPPicCn); mInfo != nil {
			if mInfo.Size < 1*1024 || mInfo.Size > 500*1024 {
				cloudStorageService.DeleteCloudMedia(ctx, smallPPicCn)
			}
		}

		if mInfo := cloudStorageService.GetMediaInfo(ctx, mediumPPicCn); mInfo != nil {
			if mInfo.Size < 1*1024 || mInfo.Size > 1*1024*1024 {
				cloudStorageService.DeleteCloudMedia(ctx, mediumPPicCn)
			}
		}

		if mInfo := cloudStorageService.GetMediaInfo(ctx, largePPicCn); mInfo != nil {
			if mInfo.Size < 1*1024 || mInfo.Size > 2*1024*1024 {
				cloudStorageService.DeleteCloudMedia(ctx, largePPicCn)
			}
		}
	}(d.PictureCloudName)

	return nil
}

type addUsersToGroupAction struct {
	NewUsers []string `msgpack:"newUsers"`
}

func (d addUsersToGroupAction) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.NewUsers, validation.Required, validation.Length(1, 0).Error("no users provided")),
	)

	return helpers.ValidationError(err, "gccValidation.go", "addUsersToGroupAction")

}

type actOnSingleUserAction struct {
	User string `msgpack:"user"`
}

func (d actOnSingleUserAction) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.User, validation.Required),
	)

	return helpers.ValidationError(err, "gccValidation.go", "actOnSingleUserAction")

}

type sendGroupChatMsg struct {
	GroupId          string               `msgpack:"groupId"`
	IsReply          bool                 `msgpack:"isReply"`
	ReplyTargetMsgId string               `msgpack:"replyTargetMsgId"`
	Msg              chatTypes.MsgContent `msgpack:"msg"`
	At               int64                `msgpack:"at"`
}

func (vb sendGroupChatMsg) Validate(ctx context.Context) error {
	err := validation.ValidateStruct(&vb,
		validation.Field(&vb.GroupId, validation.Required, is.UUID),
		validation.Field(&vb.ReplyTargetMsgId, is.UUID),
		validation.Field(&vb.Msg, validation.Required),
		validation.Field(&vb.At, validation.Required, validation.Max(time.Now().UTC().UnixMilli()).Error("invalid future time")),
	)

	return helpers.ValidationError(err, "gccValidation.go", "sendGroupChatMsg")
}

type groupChatMsgAck struct {
	GroupId string `msgpack:"groupId"`
	MsgIds  []any  `msgpack:"msgId"`
	At      int64  `msgpack:"at"`
}

func (d groupChatMsgAck) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.GroupId, validation.Required, is.UUID),
		validation.Field(&d.MsgIds, validation.Required, validation.Each(is.UUID)),
		validation.Field(&d.At, validation.Required, validation.Max(time.Now().UTC().UnixMilli()).Error("invalid future time")),
	)

	return helpers.ValidationError(err, "gccValidation.go", "groupChatMsgAck")
}

type groupInfo struct {
	GroupId string `msgpack:"groupId"`
}

func (d groupInfo) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.GroupId, validation.Required, is.UUID),
	)

	return helpers.ValidationError(err, "gccValidation.go", "groupInfo")
}
