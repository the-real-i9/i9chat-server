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
	PartnerUsername  string               `msgpack:"partnerUsername"`
	IsReply          bool                 `msgpack:"isReply"`
	ReplyTargetMsgId string               `msgpack:"replyTargetMsgId"`
	Msg              chatTypes.MsgContent `msgpack:"msg"`
	At               int64                `msgpack:"at"`
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
	MsgIds          []any  `msgpack:"msgIds"`
	PartnerUsername string `msgpack:"partnerUsername"`
	At              int64  `msgpack:"at"`
}

func (d directChatMsgsAck) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.MsgIds, validation.Required, validation.Each(is.UUID)),
		validation.Field(&d.PartnerUsername, validation.Required),
		validation.Field(&d.At, validation.Required, validation.Max(time.Now().UTC().UnixMilli()).Error("invalid future time")),
	)

	return helpers.ValidationError(err, "dccValidation.go", "directChatMsgAck")
}
