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
		validation.Field(&b.PictureData,
			validation.Required,
			validation.Length(1, 2*1024*1024).Error("maximum picture size of 2mb exceeded"),
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

type newDMChatDataT struct {
	PartnerId int            `json:"partnerId"`
	InitMsg   map[string]any `json:"initMsg"`
	CreatedAt time.Time      `json:"createdAt"`
}

func (ob newDMChatDataT) Validate() error {
	var vb struct {
		PartnerId int                 `json:"partnerId"`
		InitMsg   appTypes.MsgContent `json:"initMsg"`
		CreatedAt time.Time           `json:"createdAt"`
	}

	helpers.ToStruct(ob, &vb)

	return validation.ValidateStruct(&vb,
		validation.Field(&vb.PartnerId,
			validation.Required,
			validation.Min(1).Error("invalid value"),
		),
		validation.Field(&vb.InitMsg, validation.Required),
		validation.Field(&vb.CreatedAt,
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

type newGroupChatDataT struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	PictureData []byte              `json:"pictureData"`
	InitUsers   [][]appTypes.String `json:"initUsers"`
}

func (b newGroupChatDataT) Validate() error {
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

type ackMsgsDataT struct {
	Status      string                          `json:"status"`
	GroupChatId int                             `json:"groupChatId"`
	MsgAckDatas []*appTypes.GroupChatMsgAckData `json:"msgAckDatas"`
}

func (b ackMsgsDataT) Validate() error {
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
	LiveLocation string `json:"liveLocation"`
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
	NewGeolocation string `json:"newGeolocation"`
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
	Presence string           `json:"presence"`
	LastSeen pgtype.Timestamp `json:"lastSeen"`
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
