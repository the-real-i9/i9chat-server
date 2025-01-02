package dmChatControllers

import (
	"i9chat/appTypes"
	"i9chat/helpers"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type getChatHistoryBody struct {
	DMChatId string `json:"dmChatId"`
	Offset   int    `json:"offset"`
}

func (b getChatHistoryBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.DMChatId, validation.Required, validation.Min(1).Error("invalid value")),
		validation.Field(&b.Offset, validation.Min(0).Error("invalid negative offset")),
	)

	return helpers.ValidationError(err, "dmChat_bodyValidators.go", "getChatHistoryBody")
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

	return helpers.ValidationError(err, "dmChat_bodyValidators.go", "sendMessageBody")
}

type newDMChatBody struct {
	PartnerUserId int            `json:"partnerUserId"`
	InitMsg       map[string]any `json:"initMsg"`
	CreatedAt     time.Time      `json:"createdAt"`
}

func (b newDMChatBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.PartnerUserId,
			validation.Required,
			validation.Min(1).Error("invalid value"),
		),
		validation.Field(&b.InitMsg, validation.Required),
		validation.Field(&b.CreatedAt,
			validation.Required,
			validation.Max(time.Now()).Error("invalid future time"),
		),
	)

	return helpers.ValidationError(err, "dmChat_bodyValidators.go", "newDMChatBody")
}
