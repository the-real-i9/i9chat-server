package dmChatControllers

import (
	"i9chat/src/appTypes"
	"i9chat/src/helpers"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type getChatHistoryBody struct {
	DMChatId int `json:"dmChatId"`
	Offset   int `json:"offset"`
}

func (b getChatHistoryBody) Validate() error {
	return validation.ValidateStruct(&b,
		validation.Field(&b.DMChatId, validation.Required, validation.Min(1).Error("invalid value")),
		validation.Field(&b.Offset, validation.Min(0).Error("invalid negative offset")),
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
