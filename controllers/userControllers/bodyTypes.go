package userControllers

import (
	"i9chat/utils/appTypes"
	"i9chat/utils/helpers"
	"regexp"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/jackc/pgx/v5/pgtype"
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
		validation.Field(&b.Data, validation.Required),
	)
}

type newDMChatBodyT struct {
	PartnerId int            `json:"partnerId"`
	InitMsg   map[string]any `json:"initMsg"`
	CreatedAt time.Time      `json:"createdAt"`
}

func (b newDMChatBodyT) Validate() error {
	return validation.ValidateStruct(&b,
		validation.Field(&b.PartnerId,
			validation.Required,
			validation.Min(1).Error("invalid value"),
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

type openGroupChatStreamBody struct {
	Action string         `json:"action"`
	Data   map[string]any `json:"data"`
}

func (b openGroupChatStreamBody) Validate() error {
	return validation.ValidateStruct(&b,
		validation.Field(&b.Action,
			validation.Required,
			validation.In("create new chat", "acknowledge messages").Error("invalid action"),
		),
		validation.Field(&b.Data, validation.Required),
	)
}

type newGroupChatBodyT struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	PictureData []byte `json:"pictureData"`
	InitUsers   [][]appTypes.String
}

func (b newGroupChatBodyT) Validate() error {
	return validation.ValidateStruct(&b,
		validation.Field(&b.Name, validation.Required),
		validation.Field(&b.Description, validation.Required),
		validation.Field(&b.PictureData,
			validation.Required,
			validation.Length(1, 2*1024*1024).Error("maximum picture size of 2mb exceeded"),
		),
		validation.Field(&b.InitUsers,
			validation.Required,
			validation.Each(
				validation.Length(2, 2).Error(`invalid format; should be: [{userId}, {username}] e.g. [2, "kenny"]`),
				validation.By(helpers.UserSliceRule),
			),
		),
	)
}

type ackMsgsBodyT struct {
	Status      string                          `json:"status"`
	GroupChatId int                             `json:"groupChatId"`
	MsgAckDatas []*appTypes.GroupChatMsgAckData `json:"msgAckDatas"`
}

func (b ackMsgsBodyT) Validate() error {
	return validation.ValidateStruct(&b,
		validation.Field(&b.Status,
			validation.Required,
			validation.In("delivered", "seen").Error("invalid status value; should be 'delivered' or 'seen'"),
		),
		validation.Field(&b.GroupChatId, validation.Required, validation.Min(1).Error("invalid value")),
		validation.Field(&b.MsgAckDatas, validation.Required),
	)
}

type findNearbyUsersBody struct {
	LiveLocation string
}

func (b findNearbyUsersBody) Validate() error {
	return validation.ValidateStruct(&b,
		validation.Field(&b.LiveLocation,
			validation.Required,
			validation.Match(regexp.MustCompile("^[0-9]+, ?[0-9]+, ?[0-9]+$")).Error("invalid circle format; format: pointX, pointY, radiusR"),
		),
	)
}

type updateMyGeolocationBody struct {
	NewGeolocation string
}

func (b updateMyGeolocationBody) Validate() error {
	return validation.ValidateStruct(&b,
		validation.Field(&b.NewGeolocation,
			validation.Required,
			validation.Match(regexp.MustCompile("^[0-9]+, ?[0-9]+, ?[0-9]+$")).Error("invalid circle format; format: pointX, pointY, radiusR"),
		),
	)
}

type switchMyPresenceBody struct {
	Presence string
	LastSeen pgtype.Timestamp
}

func (b switchMyPresenceBody) Validate() error {
	return validation.ValidateStruct(&b,
		validation.Field(&b.Presence,
			validation.Required,
			validation.In("online", "offline").Error("invalid value for prescence; expects 'online' or 'offline'"),
		),
		validation.Field(&b.LastSeen,
			validation.When(b.Presence == "online", validation.Nil.Error("only required when presence is 'offline'")).Else(validation.Required, validation.Max(time.Now()).Error("invalid future time")),
		),
	)
}
