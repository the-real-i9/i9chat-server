package groupChatControllers

import (
	"i9chat/helpers"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type getChatHistoryQuery struct {
	Limit  int       `json:"limit"`
	Offset time.Time `json:"offset"`
}

func (b getChatHistoryQuery) Validate() error {

	err := validation.ValidateStruct(&b,
		validation.Field(&b.Limit, validation.Required, validation.Min(1).Error("invalid value")),
		validation.Field(&b.Offset, validation.Required, validation.Min(time.Now()).Error("invalid future time")),
	)

	return helpers.ValidationError(err, "dmChat_validation.go", "getChatHistoryQuery")
}

type newGroupChatBody struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	PictureData []byte    `json:"pictureData"`
	InitUsers   []string  `json:"initUsers"`
	CreatedAt   time.Time `json:"createdAt"`
}

func (b newGroupChatBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.Name, validation.Required),
		validation.Field(&b.Description, validation.Required),
		validation.Field(&b.PictureData,
			validation.Required,
			validation.Length(1, 2*1024*1024).Error("maximum picture size of 2mb exceeded"),
		),
		validation.Field(&b.InitUsers, validation.Required, validation.Length(1, 0).Error("at least 1 other user is required to start a group")),
		validation.Field(&b.CreatedAt, validation.Required, validation.Min(time.Now()).Error("invalid future time")),
	)

	return helpers.ValidationError(err, "groupChat_bodyValidators.go", "newGroupChatBody")

}

type action string

type executeActionParams struct {
	GroupId string `json:"groupId"`
	Action  action `json:"action"`
}

func (b executeActionParams) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.Action,
			validation.Required,
			validation.In("change name", "change description", "change picture", "add users", "remove user", "join", "leave", "make user admin", "remove user from admins").Error("invalid group action"),
		),
		validation.Field(&b.GroupId, validation.Required, is.UUID),
	)

	return helpers.ValidationError(err, "groupChat_validation.go", "executeActionParams")

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
		validation.Field(&d.NewUsers, validation.Required, validation.Length(1, 0)),
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
