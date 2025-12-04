package appTypes

import (
	"slices"

	"github.com/goccy/go-json"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type ClientUser struct {
	Username      string `json:"username"`
	ProfilePicUrl string `json:"profile_pic_url"`
}

func (c ClientUser) MarshalBinary() ([]byte, error) {
	return json.Marshal(c)
}

type BinableMap map[string]any

func (c BinableMap) MarshalBinary() ([]byte, error) {
	return json.Marshal(c)
}

type BinableSlice []any

func (c BinableSlice) MarshalBinary() ([]byte, error) {
	return json.Marshal(c)
}

type ServerEventMsg struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
}

type UserGeolocation struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type MsgProps struct {
	TextContent *string `json:"text_content,omitempty"`
	Data        []byte  `json:"data,omitempty"`
	Url         *string `json:"url,omitempty"`
	Duration    *int64  `json:"duration,omitempty"`
	Caption     *string `json:"caption,omitempty"`
	MimeType    *string `json:"mime_type,omitempty"`
	Size        *int64  `json:"size,omitempty"`
	Name        *string `json:"name,omitempty"`
	Extension   *string `json:"extension,omitempty"`
}

type MsgContent struct {
	Type      string `json:"type"`
	*MsgProps `json:"props"`
}

func (m *MsgContent) SetMediaMIME(mediaType, mediaExt string) {
	m.MimeType = &mediaType
	m.Extension = &mediaExt
}

func (m *MsgContent) SetMediaSize(mediaSize int64) {
	m.Size = &mediaSize
}

func (m *MsgContent) SetMediaUrl(mediaUrl string) {
	m.Data = nil

	m.Url = &mediaUrl
}

func (m MsgContent) Validate() error {
	msgType := m.Type

	return validation.ValidateStruct(&m,
		validation.Field(&m.Type,
			validation.Required,
			validation.In("text", "voice", "audio", "video", "photo", "file").Error("invalid message type"),
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
