package groupChatControllers

import (
	"i9chat/appTypes"
	"i9chat/helpers"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type getChatHistoryBody struct {
	GroupChatId int `json:"groupChatId"`
	Offset      int `json:"offset"`
}

func (b getChatHistoryBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.GroupChatId, validation.Required, validation.Min(1).Error("invalid value")),
		validation.Field(&b.Offset, validation.Required, validation.Min(0).Error("invalid negative offset")),
	)

	return helpers.ValidationError(err, "groupChat_bodyValidators.go", "getChatHistoryBody")

}

type sendMessageBody struct {
	Msg map[string]any `json:"msg"`
	At  time.Time      `json:"at"`
}

func (ob sendMessageBody) Validate() error {
	var vb struct {
		Msg appTypes.MsgContent `json:"msg"`
		At  time.Time           `json:"at"`
	}

	helpers.ToStruct(ob, &vb)

	err := validation.ValidateStruct(&vb,
		validation.Field(&vb.Msg, validation.Required),
		validation.Field(&vb.At,
			validation.Required,
			validation.Max(time.Now()).Error("invalid future time"),
		),
	)

	return helpers.ValidationError(err, "groupChat_bodyValidators.go", "sendMessageBody")

}

type createNewGroupChatAndAckMessagesBody struct {
	Action string         `json:"action"`
	Data   map[string]any `json:"data"`
}

func (b createNewGroupChatAndAckMessagesBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.Action,
			validation.Required,
			validation.In("create new chat", "acknowledge messages").Error("invalid action"),
		),
		validation.Field(&b.Data, validation.Required),
	)

	return helpers.ValidationError(err, "groupChat_bodyValidators.go", "createNewGroupChatAndAckMessagesBody")

}

type newGroupChatDataT struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	PictureData []byte              `json:"pictureData"`
	InitUsers   [][]appTypes.String `json:"initUsers"`
}

func (b newGroupChatDataT) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.Name, validation.Required),
		validation.Field(&b.Description, validation.Required),
		validation.Field(&b.PictureData,
			validation.Required,
			validation.Length(1, 2*1024*1024).Error("maximum picture size of 2mb exceeded"),
		),
		validation.Field(&b.InitUsers,
			validation.Required,
			validation.Each(
				validation.Length(2, 2).Error(`invalid format; should be: [{userId}, {username}] e.g. [2, "kenny"]`),
				validation.By(helpers.UserSliceRule),
			),
		),
	)

	return helpers.ValidationError(err, "groupChat_bodyValidators.go", "newGroupChatDataT")

}

type ackMsgsDataT struct {
	Status      string                          `json:"status"`
	GroupChatId int                             `json:"groupChatId"`
	MsgAckDatas []*appTypes.GroupChatMsgAckData `json:"msgAckDatas"`
}

func (b ackMsgsDataT) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.Status,
			validation.Required,
			validation.In("delivered", "seen").Error("invalid status value; should be 'delivered' or 'seen'"),
		),
		validation.Field(&b.GroupChatId, validation.Required, validation.Min(1).Error("invalid value")),
		validation.Field(&b.MsgAckDatas, validation.Required),
	)

	return helpers.ValidationError(err, "groupChat_bodyValidators.go", "ackMsgsDataT")

}

type action string

type executeActionBody struct {
	Action action         `json:"action"`
	Data   map[string]any `json:"data"`
}

func (b executeActionBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.Action,
			validation.Required,
			validation.In("change name", "change description", "change picture", "add users", "remove user", "join", "leave", "make user admin", "remove user from admins").Error("invalid group action"),
		),
		validation.Field(&b.Data, validation.Required),
	)

	return helpers.ValidationError(err, "groupChat_bodyValidators.go", "executeActionBody")

}

type changeGroupNameT struct {
	GroupChatId int    `json:"groupChatId"`
	NewName     string `json:"newName"`
}

func (d changeGroupNameT) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.GroupChatId,
			validation.Required,
			validation.Min(1).Error("invalid value"),
		),
		validation.Field(&d.NewName, validation.Required),
	)

	return helpers.ValidationError(err, "groupChat_bodyValidators.go", "changeGroupNameT")

}

type changeGroupDescriptionT struct {
	GroupChatId    int    `json:"groupChatId"`
	NewDescription string `json:"newDescription"`
}

func (d changeGroupDescriptionT) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.GroupChatId,
			validation.Required,
			validation.Min(1).Error("invalid value"),
		),
		validation.Field(&d.NewDescription, validation.Required),
	)

	return helpers.ValidationError(err, "groupChat_bodyValidators.go", "changeGroupDescriptionT")

}

type changeGroupPictureT struct {
	GroupChatId    int    `json:"groupChatId"`
	NewPictureData []byte `json:"newPictureData"`
}

func (d changeGroupPictureT) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.GroupChatId,
			validation.Required,
			validation.Min(1).Error("invalid value"),
		),
		validation.Field(&d.NewPictureData,
			validation.Required,
			validation.Length(1, 2*1024*1024).Error("maximum picture size of 2mb exceeded"),
		),
	)

	return helpers.ValidationError(err, "groupChat_bodyValidators.go", "changeGroupPictureT")

}

type addUsersToGroupT struct {
	GroupChatId int                 `json:"groupChatId"`
	NewUsers    [][]appTypes.String `json:"newUsers"`
}

func (d addUsersToGroupT) Validate() error {
	err := validation.ValidateStruct(&d,
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

	return helpers.ValidationError(err, "groupChat_bodyValidators.go", "addUsersToGroupT")

}

type actOnSingleUserT struct {
	GroupChatId int               `json:"groupChatId"`
	User        []appTypes.String `json:"user"`
}

func (d actOnSingleUserT) Validate() error {
	err := validation.ValidateStruct(&d,
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

	return helpers.ValidationError(err, "groupChat_bodyValidators.go", "actOnSingleUserT")

}

type joinLeaveGroupT struct {
	GroupChatId int `json:"groupChatId"`
}

func (d joinLeaveGroupT) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.GroupChatId,
			validation.Required,
			validation.Min(1).Error("invalid value"),
		),
	)

	return helpers.ValidationError(err, "groupChat_bodyValidators.go", "joinLeaveGroupT")

}
