package appTypes

import (
	"bytes"
	"regexp"
	"slices"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type SignupSessionData struct {
	Email                   string    `json:"email"`
	VerificationCode        int       `json:"verificationCode"`
	VerificationCodeExpires time.Time `json:"verificationCodeExpires"`
}

type SignupSession struct {
	Step string             `json:"step"`
	Data *SignupSessionData `json:"data"`
}

type ClientUser struct {
	Username string
}

type UserGeolocation struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

func (ug UserGeolocation) Validate() error {
	return validation.ValidateStruct(&ug,
		validation.Field(&ug.Longitude,
			validation.Required,
			is.Float.Error("value must be of type float"),
		),
		validation.Field(&ug.Latitude,
			validation.Required,
			is.Float.Error("value must be type float"),
		),
	)
}

type MsgProps struct {
	TextContent *string `json:"textContent,omitempty"`
	Data        []byte  `json:"data,omitempty"`
	Url         *string `json:"url,omitempty"`
	Duration    *string `json:"duration,omitempty"`
	Caption     *string `json:"caption,omitempty"`
	MimeType    *string `json:"mimeType,omitempty"`
	Size        *int    `json:"size,omitempty"`
	Name        *string `json:"name,omitempty"`
	Extension   *string `json:"extension,omitempty"`
}

type MsgContent struct {
	Type      string `json:"type"`
	*MsgProps `json:"props"`
}

func (m MsgContent) Validate() error {
	msgType := m.Type

	return validation.ValidateStruct(&m,
		validation.Field(&m.Type,
			validation.Required,
			validation.In("text", "voice", "audio", "video", "image", "file").Error("invalid message type"),
		),
		validation.Field(&m.MsgProps, validation.Required),
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
