package groupChatControllers

import (
	"i9chat/src/appTypes"
	"i9chat/src/helpers"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type getChatHistoryBody struct {
	GroupChatId int `json:"groupChatId"`
	Offset      int `json:"offset"`
}

func (b getChatHistoryBody) Validate() error {
	return validation.ValidateStruct(&b,
		validation.Field(&b.GroupChatId, validation.Required, validation.Min(1).Error("invalid value")),
		validation.Field(&b.Offset, validation.Required, validation.Min(0).Error("invalid negative offset")),
	)
}

type openMessagingStreamBody struct {
	Msg map[string]any `json:"msg"`
	At  time.Time      `json:"at"`
}

func (ob openMessagingStreamBody) Validate() error {
	var vb struct {
		Msg appTypes.MsgContent `json:"msg"`
		At  time.Time           `json:"at"`
	}

	helpers.ToStruct(ob, &vb)

	return validation.ValidateStruct(&vb,
		validation.Field(&vb.Msg, validation.Required),
		validation.Field(&vb.At,
			validation.Required,
			validation.Max(time.Now()).Error("invalid future time"),
		),
	)
}

type action string

type executeActionBody struct {
	Action action         `json:"action"`
	Data   map[string]any `json:"data"`
}

func (b executeActionBody) Validate() error {
	return validation.ValidateStruct(&b,
		validation.Field(&b.Action,
			validation.Required,
			validation.In("change name", "change description", "change picture", "add users", "remove user", "join", "leave", "make user admin", "remove user from admins").Error("invalid group action"),
		),
		validation.Field(&b.Data, validation.Required),
	)
}

type changeGroupNameT struct {
	GroupChatId int    `json:"groupChatId"`
	NewName     string `json:"newName"`
}

func (d changeGroupNameT) Validate() error {
	return validation.ValidateStruct(&d,
		validation.Field(&d.GroupChatId,
			validation.Required,
			validation.Min(1).Error("invalid value"),
		),
		validation.Field(&d.NewName, validation.Required),
	)
}

type changeGroupDescriptionT struct {
	GroupChatId    int    `json:"groupChatId"`
	NewDescription string `json:"newDescription"`
}

func (d changeGroupDescriptionT) Validate() error {
	return validation.ValidateStruct(&d,
		validation.Field(&d.GroupChatId,
			validation.Required,
			validation.Min(1).Error("invalid value"),
		),
		validation.Field(&d.NewDescription, validation.Required),
	)
}

type changeGroupPictureT struct {
	GroupChatId    int    `json:"groupChatId"`
	NewPictureData []byte `json:"newPictureData"`
}

func (d changeGroupPictureT) Validate() error {
	return validation.ValidateStruct(&d,
		validation.Field(&d.GroupChatId,
			validation.Required,
			validation.Min(1).Error("invalid value"),
		),
		validation.Field(&d.NewPictureData,
			validation.Required,
			validation.Length(1, 2*1024*1024).Error("maximum picture size of 2mb exceeded"),
		),
	)
}

type addUsersToGroupT struct {
	GroupChatId int                 `json:"groupChatId"`
	NewUsers    [][]appTypes.String `json:"newUsers"`
}

func (d addUsersToGroupT) Validate() error {
	return validation.ValidateStruct(&d,
		validation.Field(&d.GroupChatId,
			validation.Required,
			validation.Min(1).Error("invalid value"),
		),
		validation.Field(&d.NewUsers,
			validation.Required,
			validation.Each(
				validation.Length(2, 2).Error(`invalid format; should be: [{userId}, {username}] e.g. [2, "kenny"]`),
				validation.By(helpers.UserSliceRule),
			),
		),
	)
}

type actOnSingleUserT struct {
	GroupChatId int               `json:"groupChatId"`
	User        []appTypes.String `json:"user"`
}

func (d actOnSingleUserT) Validate() error {
	return validation.ValidateStruct(&d,
		validation.Field(&d.GroupChatId,
			validation.Required,
			validation.Min(1).Error("invalid value"),
		),
		validation.Field(&d.User,
			validation.Required,
			validation.Length(2, 2).Error(`invalid format; should be: [{userId}, {username}] e.g. [2, "kenny"]`),
			validation.By(helpers.UserSliceRule),
		),
	)
}

type joinLeaveGroupT struct {
	GroupChatId int `json:"groupChatId"`
}

func (d joinLeaveGroupT) Validate() error {
	return validation.ValidateStruct(&d,
		validation.Field(&d.GroupChatId,
			validation.Required,
			validation.Min(1).Error("invalid value"),
		),
	)
}
