package appTypes

import (
	"bytes"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type SignupSessionData struct {
	SessionId string
	Email     string
}

type ClientUser struct {
	Id       int
	Username string
}

type DMChatMsgAckData struct {
	MsgId    int       `json:"msgId"`
	DMChatId int       `json:"dmChatId"`
	SenderId int       `json:"senderId"`
	At       time.Time `json:"at"`
}

func (d DMChatMsgAckData) Validate() error {
	return validation.ValidateStruct(&d,
		validation.Field(&d.MsgId, validation.Required, validation.Min(1).Error("invalid value")),
		validation.Field(&d.DMChatId, validation.Required, validation.Min(1).Error("invalid value")),
		validation.Field(&d.SenderId, validation.Required, validation.Min(1).Error("invalid value")),
		validation.Field(&d.At, validation.Required, validation.Max(time.Now()).Error("invalid future time")),
	)
}

type GroupChatMsgAckData struct {
	MsgId int       `json:"msgId"`
	At    time.Time `json:"at"`
}

func (d GroupChatMsgAckData) Validate() error {
	return validation.ValidateStruct(&d,
		validation.Field(&d.MsgId, validation.Required, validation.Min(1).Error("invalid value")),
		validation.Field(&d.At, validation.Required, validation.Max(time.Now()).Error("invalid future time")),
	)
}

type WSResp struct {
	StatusCode int    `json:"statusCode"`
	Body       any    `json:"body"`
	Error      string `json:"error"`
}

type String string

func (s *String) UnmarshalJSON(b []byte) error {
	nb := bytes.Trim(b, "\"")

	*s = String(nb)

	return nil
}
