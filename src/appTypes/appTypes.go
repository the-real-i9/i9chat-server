package appTypes

import (
	"slices"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type ClientUser struct {
	Username string
}

type ServerWSMsg struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
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
	Duration    *int64  `json:"duration,omitempty"`
	Caption     *string `json:"caption,omitempty"`
	MimeType    *string `json:"mimeType,omitempty"`
	Size        *int64  `json:"size,omitempty"`
	Name        *string `json:"name,omitempty"`
	Extension   *string `json:"extension,omitempty"`
}

type MsgContent struct {
	Type      string `json:"type"`
	*MsgProps `json:"props"`
}

func (m *MsgContent) SetMediaMIME(mediaType, mediaExt string) {
	mediaTypePtr := new(string)
	*mediaTypePtr = mediaType

	mediaExtPtr := new(string)
	*mediaExtPtr = mediaExt

	m.MimeType = mediaTypePtr
	m.Extension = mediaExtPtr
}

func (m *MsgContent) SetMediaSize(size int64) {
	mediaSizePtr := new(int64)
	*mediaSizePtr = size

	m.Size = mediaSizePtr
}

func (m *MsgContent) SetMediaUrl(url string) {
	m.Data = nil

	mediaUrlPtr := new(string)
	*mediaUrlPtr = url

	m.Url = mediaUrlPtr
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
		validation.Field(&m.Data,
			validation.When(msgType == "text", validation.Nil.Error("invalid property for the specified type")).Else(
				validation.Required,
				validation.Length(1024, 10*1024*1024).Error("data size oute of range. min: 1KiB, max: 10MiB"),
			),
		),
		validation.Field(&m.Duration, validation.When(msgType != "voice", validation.Nil.Error("invalid property for the specified type")).Else(validation.Required)),
		validation.Field(&m.Caption, validation.When(slices.Contains([]string{"text", "voice", "file", "audio"}, msgType), validation.Nil.Error("invalid property for the specified type")).Else(validation.Required)),
		validation.Field(&m.Name, validation.When(msgType != "file", validation.Nil.Error("invalid property for the specified type")).Else(validation.Required)),

		validation.Field(&m.Size, validation.Nil.Error("setting this property is forbidden")),
		validation.Field(&m.Url, validation.Nil.Error("setting this property is forbidden")),
		validation.Field(&m.MimeType, validation.Nil.Error("setting this property is forbidden")),
		validation.Field(&m.Extension, validation.Nil.Error("setting this property is forbidden")),
	)
}
