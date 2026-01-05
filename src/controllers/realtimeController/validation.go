package realtimeController

import (
	"context"
	"errors"
	"fmt"
	"i9chat/src/appGlobals"
	"i9chat/src/helpers"
	"os"
	"slices"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/gofiber/fiber/v2"
)

// Realtime Action Body
type rtActionBody struct {
	Action string         `json:"action"`
	Data   map[string]any `json:"data"`
}

func (b rtActionBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.Action, validation.Required),
		validation.Field(&b.Data, validation.Required),
	)

	return helpers.ValidationError(err, "realtimeController_validation.go", "rtActionBody")
}

type subToUserPresenceAcd struct {
	Usernames []string `json:"users"`
}

func (vb subToUserPresenceAcd) Validate() error {
	err := validation.ValidateStruct(&vb,
		validation.Field(&vb.Usernames, validation.Required, validation.Length(1, 0)),
	)

	return helpers.ValidationError(err, "realtimeController_validation.go", "subToUserPresenceAcd")
}

type unsubFromUserPresenceAcd struct {
	Usernames []string `json:"users"`
}

func (vb unsubFromUserPresenceAcd) Validate() error {
	err := validation.ValidateStruct(&vb,
		validation.Field(&vb.Usernames, validation.Required, validation.Length(1, 0)),
	)

	return helpers.ValidationError(err, "realtimeController_validation.go", "unsubFromUserPresenceAcd")
}

type MsgProps struct {
	TextContent    *string `json:"text_content,omitempty"`
	MediaCloudName *string `json:"media_cloud_name,omitempty"`
	Duration       *int64  `json:"duration,omitempty"`
	Caption        *string `json:"caption,omitempty"`
	Name           *string `json:"name,omitempty"`

	// fields to set when sending to client
	// Url            *string `json:"url,omitempty"`
	// MimeType       *string `json:"mime_type,omitempty"`
	// Size           *int64  `json:"size,omitempty"`
	// Extension      *string `json:"extension,omitempty"`
}

type MsgContent struct {
	Type      string `json:"type"`
	*MsgProps `json:"props"`
}

func (m MsgContent) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Type,
			validation.Required,
			validation.In("text", "voice", "audio", "video", "photo", "file").Error("invalid message type"),
		),
		validation.Field(&m.MediaCloudName,
			validation.When(m.Type == "text", validation.Nil.Error("invalid property for the specified type")).Else(
				validation.Required,
				validation.By(func(value any) error {
					val := value.(string)

					if !(strings.HasPrefix(val, "blur_placeholder:uploads/chat/") && strings.Contains(val, " actual:uploads/chat/")) {
						return errors.New("invalid media cloud name")
					}

					return nil
				}),
			),
		),
		validation.Field(&m.MsgProps, validation.Required),
		validation.Field(&m.TextContent, validation.When(m.Type != "text", validation.Nil.Error("invalid property for the specified type")).Else(validation.Required)),
		validation.Field(&m.Duration, validation.When(m.Type != "voice", validation.Nil.Error("invalid property for the specified type")).Else(validation.Required)),
		validation.Field(&m.Caption, validation.When(slices.Contains([]string{"text", "voice", "file", "audio"}, m.Type), validation.Nil.Error("invalid property for the specified type")).Else(validation.Required)),
		validation.Field(&m.Name, validation.When(m.Type != "file", validation.Nil.Error("invalid property for the specified type")).Else(validation.Required)),
	)
}

type sendDirectChatMsg struct {
	PartnerUsername  string      `json:"partnerUsername"`
	IsReply          bool        `json:"isReply"`
	ReplyTargetMsgId string      `json:"replyTargetMsgId"`
	Msg              *MsgContent `json:"msg"`
	At               int64       `json:"at"`
}

func (vb sendDirectChatMsg) Validate(ctx context.Context) error {
	err := validation.ValidateStruct(&vb,
		validation.Field(&vb.PartnerUsername, validation.Required),
		validation.Field(&vb.ReplyTargetMsgId, is.UUID),
		validation.Field(&vb.Msg, validation.Required),
		validation.Field(&vb.At, validation.Required, validation.Max(time.Now().UTC().UnixMilli()).Error("invalid future time")),
	)

	if err != nil {
		return helpers.ValidationError(err, "realtimeController_validation.go", "sendDirectChatMsg")
	}

	if mediaCloudName := *vb.Msg.MediaCloudName; mediaCloudName != "" {
		_, err = appGlobals.GCSClient.Bucket(os.Getenv("GCS_BUCKET_NAME")).Object(mediaCloudName).Attrs(ctx)
		if errors.Is(err, storage.ErrObjectNotExist) {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("upload error: media (%s) does not exist in cloud", mediaCloudName))
		}
	}

	return nil
}

type directChatMsgAck struct {
	MsgId           string `json:"msgId"`
	PartnerUsername string `json:"partnerUsername"`
	At              int64  `json:"at"`
}

func (d directChatMsgAck) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.MsgId, validation.Required, is.UUID),
		validation.Field(&d.PartnerUsername, validation.Required),
		validation.Field(&d.At, validation.Required, validation.Max(time.Now().UTC().UnixMilli()).Error("invalid future time")),
	)

	return helpers.ValidationError(err, "realtimeController_validation.go", "directChatMsgAck")
}

type sendGroupChatMsg struct {
	GroupId          string      `json:"groupId"`
	IsReply          bool        `json:"isReply"`
	ReplyTargetMsgId string      `json:"replyTargetMsgId"`
	Msg              *MsgContent `json:"msg"`
	At               int64       `json:"at"`
}

func (vb sendGroupChatMsg) Validate(ctx context.Context) error {
	err := validation.ValidateStruct(&vb,
		validation.Field(&vb.GroupId, validation.Required, is.UUID),
		validation.Field(&vb.ReplyTargetMsgId, is.UUID),
		validation.Field(&vb.Msg, validation.Required),
		validation.Field(&vb.At, validation.Required, validation.Max(time.Now().UTC().UnixMilli()).Error("invalid future time")),
	)

	if err != nil {
		return helpers.ValidationError(err, "realtimeController_validation.go", "sendGroupChatMsg")
	}

	if mediaCloudName := *vb.Msg.MediaCloudName; mediaCloudName != "" {
		_, err = appGlobals.GCSClient.Bucket(os.Getenv("GCS_BUCKET_NAME")).Object(mediaCloudName).Attrs(ctx)
		if errors.Is(err, storage.ErrObjectNotExist) {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("upload error: media (%s) does not exist in cloud", mediaCloudName))
		}
	}

	return nil
}

type groupChatMsgAck struct {
	GroupId string `json:"groupId"`
	MsgId   string `json:"msgId"`
	At      int64  `json:"at"`
}

func (d groupChatMsgAck) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.GroupId, validation.Required, is.UUID),
		validation.Field(&d.MsgId, validation.Required, is.UUID),
		validation.Field(&d.At, validation.Required, validation.Max(time.Now().UTC().UnixMilli()).Error("invalid future time")),
	)

	return helpers.ValidationError(err, "realtimeController_validation.go", "groupChatMsgAck")
}

type groupInfo struct {
	GroupId string `json:"groupId"`
}

func (d groupInfo) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.GroupId, validation.Required, is.UUID),
	)

	return helpers.ValidationError(err, "realtimeController_validation.go", "groupInfo")
}
