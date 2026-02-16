package directChatControllers

import (
	"context"
	"i9chat/src/controllers/chatControllers/chatTypes"
	"i9chat/src/helpers"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type sendDirectChatMsg struct {
	PartnerUsername  string               `json:"partnerUsername"`
	IsReply          bool                 `json:"isReply"`
	ReplyTargetMsgId string               `json:"replyTargetMsgId"`
	Msg              chatTypes.MsgContent `json:"msg"`
	At               int64                `json:"at"`
}

func (vb sendDirectChatMsg) Validate(ctx context.Context) error {
	err := validation.ValidateStruct(&vb,
		validation.Field(&vb.PartnerUsername, validation.Required),
		validation.Field(&vb.ReplyTargetMsgId, is.UUID),
		validation.Field(&vb.Msg, validation.Required),
		validation.Field(&vb.At, validation.Required, validation.Max(time.Now().UTC().UnixMilli()).Error("invalid future time")),
	)

	return helpers.ValidationError(err, "dccValidation.go", "sendDirectChatMsg")
}

type directChatMsgsAck struct {
	MsgIds          []any  `json:"msgIds"`
	PartnerUsername string `json:"partnerUsername"`
	At              int64  `json:"at"`
}

func (d directChatMsgsAck) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.MsgIds, validation.Required, validation.Each(is.UUID)),
		validation.Field(&d.PartnerUsername, validation.Required),
		validation.Field(&d.At, validation.Required, validation.Max(time.Now().UTC().UnixMilli()).Error("invalid future time")),
	)

	return helpers.ValidationError(err, "dccValidation.go", "directChatMsgAck")
}
