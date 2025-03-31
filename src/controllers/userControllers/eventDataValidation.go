package userControllers

import (
	"i9chat/src/appTypes"
	"i9chat/src/helpers"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type clientEventBody struct {
	Event string         `json:"event"`
	Data  map[string]any `json:"data"`
}

func (b clientEventBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.Event, validation.Required),
		validation.Field(&b.Data, validation.Required),
	)

	return helpers.ValidationError(err, "userControllers_validation.go", "clientEventBody")
}

type newDMChatMsg struct {
	PartnerUsername string               `json:"partnerUsername"`
	Msg             *appTypes.MsgContent `json:"msg"`
	CreatedAt       time.Time            `json:"createdAt"`
}

func (vb newDMChatMsg) Validate() error {
	err := validation.ValidateStruct(&vb,
		validation.Field(&vb.PartnerUsername, validation.Required),
		validation.Field(&vb.Msg, validation.Required),
		validation.Field(&vb.CreatedAt,
			validation.Required,
			validation.Max(time.Now()).Error("invalid future time"),
		),
	)

	return helpers.ValidationError(err, "userControllers_validation.go", "newDMChatMsg")
}

type dmChatMsgAck struct {
	MsgId           string    `json:"msgId"`
	PartnerUsername string    `json:"partnerUsername"`
	At              time.Time `json:"at"`
}

func (d dmChatMsgAck) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.MsgId, validation.Required, is.UUID),
		validation.Field(&d.PartnerUsername, validation.Required),
		validation.Field(&d.At, validation.Required, validation.Max(time.Now()).Error("invalid future time")),
	)

	return helpers.ValidationError(err, "userControllers_validation.go", "dmChatMsgAck")
}

type newGroupChatMsg struct {
	GroupId   string               `json:"groupId"`
	Msg       *appTypes.MsgContent `json:"msg"`
	CreatedAt time.Time            `json:"createdAt"`
}

func (vb newGroupChatMsg) Validate() error {
	err := validation.ValidateStruct(&vb,
		validation.Field(&vb.Msg, validation.Required),
		validation.Field(&vb.GroupId, validation.Required, is.UUID),
		validation.Field(&vb.CreatedAt, validation.Required, validation.Max(time.Now()).Error("invalid future time")),
	)

	return helpers.ValidationError(err, "userControllers_validation.go", "newGroupChatMsg")
}

type groupChatMsgAck struct {
	GroupId string    `json:"groupId"`
	MsgId   string    `json:"msgId"`
	At      time.Time `json:"at"`
}

func (d groupChatMsgAck) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.GroupId, validation.Required, is.UUID),
		validation.Field(&d.MsgId, validation.Required, is.UUID),
		validation.Field(&d.At, validation.Required, validation.Max(time.Now()).Error("invalid future time")),
	)

	return helpers.ValidationError(err, "userControllers_validation.go", "groupChatMsgAck")
}

type groupInfo struct {
	GroupId string `json:"groupId"`
}

func (d groupInfo) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.GroupId, validation.Required, is.UUID),
	)

	return helpers.ValidationError(err, "userControllers_validation.go", "groupInfo")
}

type groupMemInfo struct {
	GroupId string `json:"groupId"`
}

func (d groupMemInfo) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.GroupId, validation.Required, is.UUID),
	)

	return helpers.ValidationError(err, "userControllers_validation.go", "groupMemInfo")
}

type dmChatHistory struct {
	PartnerUsername string    `json:"partnerUsername"`
	Limit           int       `json:"limit"`
	Offset          time.Time `json:"offset"`
}

func (b dmChatHistory) Validate() error {

	err := validation.ValidateStruct(&b,
		validation.Field(&b.PartnerUsername, validation.Required),
		validation.Field(&b.Limit, validation.Min(1).Error("invalid value")),
		validation.Field(&b.Offset, validation.Min(time.Now()).Error("invalid future time")),
	)

	return helpers.ValidationError(err, "dmChat_validation.go", "dmChatHistory")
}

type groupChatHistory struct {
	GroupId string    `json:"groupId"`
	Limit   int       `json:"limit"`
	Offset  time.Time `json:"offset"`
}

func (b groupChatHistory) Validate() error {

	err := validation.ValidateStruct(&b,
		validation.Field(&b.GroupId, validation.Required, is.UUID),
		validation.Field(&b.Limit, validation.Min(1).Error("invalid value")),
		validation.Field(&b.Offset, validation.Min(time.Now()).Error("invalid future time")),
	)

	return helpers.ValidationError(err, "dmChat_validation.go", "groupChatHistory")
}
