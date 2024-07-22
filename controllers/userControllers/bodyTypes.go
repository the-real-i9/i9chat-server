package userControllers

import (
	"i9chat/utils/appTypes"
	"i9chat/utils/helpers"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type changeProfilePictureBody struct {
	PictureData []byte `json:"pictureData"`
}

func (b changeProfilePictureBody) Validate() error {
	return validation.ValidateStruct(&b,
		validation.Field(&b.PictureData,
			validation.Required,
		),
	)
}

type openDMChatStreamBody struct {
	Action string         `json:"action"`
	Data   map[string]any `json:"data"`
}

func (b openDMChatStreamBody) Validate() error {
	return validation.ValidateStruct(&b,
		validation.Field(&b.Action,
			validation.Required,
			validation.In("create new chat", "acknowledge message", "batch acknowledge messages").Error("invalid action"),
		),
		validation.Field(&b.Data,
			validation.Required,
		),
	)
}

type newChatBodyT struct {
	PartnerId int            `json:"partnerId"`
	InitMsg   map[string]any `json:"initMsg"`
	CreatedAt time.Time      `json:"createdAt"`
}

func (b newChatBodyT) Validate() error {
	return validation.ValidateStruct(&b,
		validation.Field(&b.PartnerId,
			validation.Required,
			validation.Min(0).Error("invalid negative value"),
		),
		validation.Field(&b.InitMsg, helpers.MsgContentRule(b.InitMsg["type"].(string))...),
		validation.Field(&b.CreatedAt,
			validation.Required,
			validation.Max(time.Now()).Error("invalid future time"),
		),
	)
}

type ackMsgBodyT struct {
	Status string `json:"status"`
	*appTypes.DMChatMsgAckData
}

func (b ackMsgBodyT) Validate() error {
	return validation.ValidateStruct(&b,
		validation.Field(&b.Status,
			validation.Required,
			validation.In("delivered", "seen").Error("invalid status value; should be 'delivered' or 'seen'"),
		),
		validation.Field(&b.DMChatMsgAckData),
	)
}

type batchAckMsgBodyT struct {
	Status      string                       `json:"status"`
	MsgAckDatas []*appTypes.DMChatMsgAckData `json:"msgAckDatas"`
}

func (b batchAckMsgBodyT) Validate() error {
	return validation.ValidateStruct(&b,
		validation.Field(&b.Status,
			validation.Required,
			validation.In("delivered", "seen").Error("invalid status value; should be 'delivered' or 'seen'"),
		),
		validation.Field(&b.MsgAckDatas, validation.Required),
	)
}
