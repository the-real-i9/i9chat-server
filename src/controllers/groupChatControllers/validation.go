package groupChatControllers

import (
	"i9chat/src/helpers"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
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

type changeGroupPictureAction struct {
	NewPictureData []byte `json:"newPictureData"`
}

func (d changeGroupPictureAction) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.NewPictureData,
			validation.Required,
			validation.Length(1, 2*1024*1024).Error("maximum picture size of 2mb exceeded"),
		),
	)

	return helpers.ValidationError(err, "groupChat_validation.go", "changeGroupPictureAction")

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
