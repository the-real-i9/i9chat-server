package dmChatControllers

import (
	"i9chat/utils/helpers"
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
		validation.Field(&b.Offset, validation.Required, validation.Min(0).Error("invalid negative offset")),
	)
}

type openMessagingStreamBody struct {
	Msg map[string]any `json:"msg"`
	At  time.Time      `json:"at"`
}

func (b openMessagingStreamBody) Validate() error {
	return validation.ValidateStruct(&b,
		validation.Field(&b.Msg, helpers.MsgContentRule(b.Msg["type"].(string))...),
		validation.Field(&b.At,
			validation.Required,
			validation.Max(time.Now()).Error("invalid future time"),
		),
	)
}
