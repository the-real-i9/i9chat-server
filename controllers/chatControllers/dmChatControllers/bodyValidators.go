package dmChatControllers

import (
	"i9chat/appTypes"
	"i9chat/helpers"
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

	return validation.ValidateStruct(&vb,
		validation.Field(&vb.Msg, validation.Required),
		validation.Field(&vb.At,
			validation.Required,
			validation.Max(time.Now()).Error("invalid future time"),
		),
	)
}

type createNewDMChatAndAckMessagesBody struct {
	Action string         `json:"action"`
	Data   map[string]any `json:"data"`
}

func (b createNewDMChatAndAckMessagesBody) Validate() error {
	return validation.ValidateStruct(&b,
		validation.Field(&b.Action,
			validation.Required,
			validation.In("create new chat", "acknowledge message", "batch acknowledge messages").Error("invalid action"),
		),
		validation.Field(&b.Data, validation.Required),
	)
}

type newDMChatDataT struct {
	PartnerId int            `json:"partnerId"`
	InitMsg   map[string]any `json:"initMsg"`
	CreatedAt time.Time      `json:"createdAt"`
}

func (b newDMChatDataT) Validate() error {
	return validation.ValidateStruct(&b,
		validation.Field(&b.PartnerId,
			validation.Required,
			validation.Min(1).Error("invalid value"),
		),
		validation.Field(&b.InitMsg, validation.Required),
		validation.Field(&b.CreatedAt,
			validation.Required,
			validation.Max(time.Now()).Error("invalid future time"),
		),
	)
}

type ackMsgDataT struct {
	Status string `json:"status"`
	*appTypes.DMChatMsgAckData
}

func (b ackMsgDataT) Validate() error {
	return validation.ValidateStruct(&b,
		validation.Field(&b.Status,
			validation.Required,
			validation.In("delivered", "seen").Error("invalid status value; should be 'delivered' or 'seen'"),
		),
		validation.Field(&b.DMChatMsgAckData),
	)
}

type batchAckMsgDataT struct {
	Status      string                       `json:"status"`
	MsgAckDatas []*appTypes.DMChatMsgAckData `json:"msgAckDatas"`
}

func (b batchAckMsgDataT) Validate() error {
	return validation.ValidateStruct(&b,
		validation.Field(&b.Status,
			validation.Required,
			validation.In("delivered", "seen").Error("invalid status value; should be 'delivered' or 'seen'"),
		),
		validation.Field(&b.MsgAckDatas, validation.Required),
	)
}
