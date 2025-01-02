package appTypes

import (
	"bytes"
	"regexp"
	"slices"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type SignupSessionData struct {
	Email                   string    `json:"email"`
	VerificationCode        int       `json:"verificationCode"`
	VerificationCodeExpires time.Time `json:"verificationCodeExpires"`
}

type SignupSession struct {
	Step string `json:"step"`
	Data *SignupSessionData
}

type ClientUser struct {
	Id       int
	Username string
}

type Props struct {
	TextContent *string `json:"textContent"`
	Data        []*byte `json:"data"`
	Duration    *string `json:"duration"`
	Caption     *string `json:"caption"`
	MimeType    *string `json:"mimeType"`
	Size        *int    `json:"size"`
	Name        *string `json:"name"`
	Extension   *string `json:"extension"`
}

type MsgContent struct {
	Type  string `json:"type"`
	Props `json:"props"`
}

func (m MsgContent) Validate() error {
	msgType := m.Type

	return validation.ValidateStruct(&m,
		validation.Field(&m.Type,
			validation.Required,
			validation.In("text", "voice", "audio", "video", "image", "file").Error("invalid message type"),
		),
		validation.Field(&m.Props, validation.Required),
		validation.Field(&m.TextContent, validation.When(msgType != "text", validation.Nil.Error("invalid property for the specified type")).Else(validation.Required)),
		validation.Field(&m.Data, validation.When(msgType == "text", validation.Nil.Error("invalid property for the specified type")).Else(
			validation.Required,
			validation.Length(1, 10*1024*1024).Error("maximum data size of 10mb exceeded"),
		),
		),
		validation.Field(&m.Duration, validation.When(msgType != "voice", validation.Nil.Error("invalid property for the specified type")).Else(validation.Required)),
		validation.Field(&m.MimeType,
			validation.When(slices.Contains([]string{"voice", "text"}, msgType), validation.Nil.Error("invalid property for the specified type")).Else(
				validation.Required,
				validation.Match(regexp.MustCompile("^[[:alnum:]!#$&^_.+-]+/[[:alnum:]!#$&^_.+-]+(?:;[[:blank:]]*[[:alnum:]!#$&^_.+-]+=[[:alnum:]!#$&^_.+-]+)*$")),
			),
		),
		validation.Field(&m.Size,
			validation.When(slices.Contains([]string{"voice", "text"}, msgType), validation.Nil.Error("invalid property for the specified type")).Else(
				validation.Required,
				validation.Min(1).Error("size cannot be zero bytes"),
				validation.Max(10*1024*1024).Error("maximum bytes of 10mb exceeded"),
			),
		),
		validation.Field(&m.Caption, validation.When(slices.Contains([]string{"text", "voice", "file"}, msgType), validation.Nil.Error("invalid property for the specified type")).Else(validation.Required)),
		validation.Field(&m.Extension, validation.When(msgType != "file", validation.Nil.Error("invalid property for the specified type")).Else(validation.Required)),
		validation.Field(&m.Name, validation.When(msgType != "file", validation.Nil.Error("invalid property for the specified type")).Else(validation.Required)),
	)
}

type DMChatMsgAckData struct {
	MsgId          int       `json:"msgId"`
	ClientDMChatId string    `json:"dmChatId"`
	At             time.Time `json:"at"`
}

func (d DMChatMsgAckData) Validate() error {
	return validation.ValidateStruct(&d,
		validation.Field(&d.MsgId, validation.Required, validation.Min(1).Error("invalid value")),
		validation.Field(&d.ClientDMChatId, validation.Required, validation.Min(1).Error("invalid value")),
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
