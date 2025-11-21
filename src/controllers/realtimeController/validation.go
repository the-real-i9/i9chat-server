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

type subToUserPresenceAcd struct {
	Usernames []string `json:"users"`
}

func (vb subToUserPresenceAcd) Validate() error {
	err := validation.ValidateStruct(&vb,
		validation.Field(&vb.Usernames, validation.Required, validation.Length(1, 0)),
	)

	return helpers.ValidationError(err, "realtimeController_validation.go", "subToUserPresenceAcd")
}

type unsubFromUserPresenceAcd struct {
	Usernames []string `json:"users"`
}

func (vb unsubFromUserPresenceAcd) Validate() error {
	err := validation.ValidateStruct(&vb,
		validation.Field(&vb.Usernames, validation.Required, validation.Length(1, 0)),
	)

	return helpers.ValidationError(err, "realtimeController_validation.go", "unsubFromUserPresenceAcd")
}

type sendDirectChatMsg struct {
	PartnerUsername  string               `json:"partnerUsername"`
	IsReply          bool                 `json:"isReply"`
	ReplyTargetMsgId string               `json:"replyTargetMsgId"`
	Msg              *appTypes.MsgContent `json:"msg"`
	At               int64                `json:"at"`
}

func (vb sendDirectChatMsg) Validate() error {
	err := validation.ValidateStruct(&vb,
		validation.Field(&vb.PartnerUsername, validation.Required),
		validation.Field(&vb.ReplyTargetMsgId, is.UUID),
		validation.Field(&vb.Msg, validation.Required),
		validation.Field(&vb.At,
			validation.Required,
			validation.Max(time.Now().UTC().UnixMilli()).Error("invalid future time"),
		),
	)

	return helpers.ValidationError(err, "realtimeController_validation.go", "sendDirectChatMsg")
}

type directChatMsgAck struct {
	MsgId           string `json:"msgId"`
	PartnerUsername string `json:"partnerUsername"`
	At              int64  `json:"at"`
}

func (d directChatMsgAck) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.MsgId, validation.Required, is.UUID),
		validation.Field(&d.PartnerUsername, validation.Required),
		validation.Field(&d.At, validation.Required, validation.Max(time.Now().UTC().UnixMilli()).Error("invalid future time")),
	)

	return helpers.ValidationError(err, "realtimeController_validation.go", "directChatMsgAck")
}

type sendGroupChatMsg struct {
	GroupId          string               `json:"groupId"`
	IsReply          bool                 `json:"isReply"`
	ReplyTargetMsgId string               `json:"replyTargetMsgId"`
	Msg              *appTypes.MsgContent `json:"msg"`
	At               int64                `json:"at"`
}

func (vb sendGroupChatMsg) Validate() error {
	err := validation.ValidateStruct(&vb,
		validation.Field(&vb.GroupId, validation.Required, is.UUID),
		validation.Field(&vb.ReplyTargetMsgId, is.UUID),
		validation.Field(&vb.Msg, validation.Required),
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
