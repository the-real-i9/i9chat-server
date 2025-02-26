package appTypes

import (
	"regexp"
	"slices"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type ClientUser struct {
	Username string
}

type UserGeolocation struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

func (ug UserGeolocation) Validate() error {
	return validation.ValidateStruct(&ug,
		validation.Field(&ug.X, validation.Required),
		validation.Field(&ug.Y, validation.Required),
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
