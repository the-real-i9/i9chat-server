package realtimeController

import (
	"i9chat/src/appTypes"
	"i9chat/src/helpers"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

// Realtime Action Body
type rtActionBody struct {
	Action string         `json:"action"`
	Data   map[string]any `json:"data"`
}

func (b rtActionBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.Action, validation.Required),
		validation.Field(&b.Data, validation.Required),
	)

	return helpers.ValidationError(err, "realtimeController_validation.go", "rtActionBody")
}

type sendDMChatMsg struct {
	PartnerUsername string               `json:"partnerUsername"`
	Msg             *appTypes.MsgContent `json:"msg"`
	At              int64                `json:"at"`
}

func (vb sendDMChatMsg) Validate() error {
	err := validation.ValidateStruct(&vb,
		validation.Field(&vb.PartnerUsername, validation.Required),
		validation.Field(&vb.Msg, validation.Required),
		validation.Field(&vb.At,
			validation.Required,
			validation.Max(time.Now().UTC().UnixMilli()).Error("invalid future time"),
		),
	)

	return helpers.ValidationError(err, "realtimeController_validation.go", "sendDMChatMsg")
}

type dmChatMsgAck struct {
	MsgId           string `json:"msgId"`
	PartnerUsername string `json:"partnerUsername"`
	At              int64  `json:"at"`
}

func (d dmChatMsgAck) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.MsgId, validation.Required, is.UUID),
		validation.Field(&d.PartnerUsername, validation.Required),
		validation.Field(&d.At, validation.Required, validation.Max(time.Now().UTC().UnixMilli()).Error("invalid future time")),
	)

	return helpers.ValidationError(err, "realtimeController_validation.go", "dmChatMsgAck")
}

type sendGroupChatMsg struct {
	GroupId string               `json:"groupId"`
	Msg     *appTypes.MsgContent `json:"msg"`
	At      int64                `json:"at"`
}

func (vb sendGroupChatMsg) Validate() error {
	err := validation.ValidateStruct(&vb,
		validation.Field(&vb.Msg, validation.Required),
		validation.Field(&vb.GroupId, validation.Required, is.UUID),
		validation.Field(&vb.At, validation.Required, validation.Max(time.Now().UTC().UnixMilli()).Error("invalid future time")),
	)

	return helpers.ValidationError(err, "realtimeController_validation.go", "sendGroupChatMsg")
}

type groupChatMsgAck struct {
	GroupId string `json:"groupId"`
	MsgId   string `json:"msgId"`
	At      int64  `json:"at"`
}

func (d groupChatMsgAck) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.GroupId, validation.Required, is.UUID),
		validation.Field(&d.MsgId, validation.Required, is.UUID),
		validation.Field(&d.At, validation.Required, validation.Max(time.Now().UTC().UnixMilli()).Error("invalid future time")),
	)

	return helpers.ValidationError(err, "realtimeController_validation.go", "groupChatMsgAck")
}

type groupInfo struct {
	GroupId string `json:"groupId"`
}

func (d groupInfo) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.GroupId, validation.Required, is.UUID),
	)

	return helpers.ValidationError(err, "realtimeController_validation.go", "groupInfo")
}
