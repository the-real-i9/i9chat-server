package userControllers

import (
	"i9chat/appTypes"
	"i9chat/helpers"
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type clientEventBody struct {
	Action string         `json:"action"`
	Data   map[string]any `json:"data"`
}

func (b clientEventBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.Action,
			validation.Required,
			validation.In("ACK dm chat message", "batch ACK dm chat messages", "batch ACK group chat messages").Error("invalid action"),
		),
		validation.Field(&b.Data, validation.Required),
	)

	return helpers.ValidationError(err, "user_bodyValidators.go", "clientEventBody")
}

type dmChatMsgAckDataT struct {
	Status string `json:"status"`
	*appTypes.DMChatMsgAckData
}

func (b dmChatMsgAckDataT) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.Status,
			validation.Required,
			validation.In("delivered", "seen").Error("invalid status value; should be 'delivered' or 'seen'"),
		),
		validation.Field(&b.DMChatMsgAckData),
	)

	return helpers.ValidationError(err, "user_bodyValidators.go", "dmChatMsgAckDataT")
}

type batchDMChatMsgAckDataT struct {
	Status      string                       `json:"status"`
	MsgAckDatas []*appTypes.DMChatMsgAckData `json:"msgAckDatas"`
}

func (b batchDMChatMsgAckDataT) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.Status,
			validation.Required,
			validation.In("delivered", "seen").Error("invalid status value; should be 'delivered' or 'seen'"),
		),
		validation.Field(&b.MsgAckDatas, validation.Required),
	)

	return helpers.ValidationError(err, "user_bodyValidators.go", "batchDMChatMsgAckDataT")

}

type groupChatMsgsAckDataT struct {
	Status      string                          `json:"status"`
	GroupChatId int                             `json:"groupChatId"`
	MsgAckDatas []*appTypes.GroupChatMsgAckData `json:"msgAckDatas"`
}

func (b groupChatMsgsAckDataT) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.Status,
			validation.Required,
			validation.In("delivered", "seen").Error("invalid status value; should be 'delivered' or 'seen'"),
		),
		validation.Field(&b.GroupChatId, validation.Required, validation.Min(1).Error("invalid value")),
		validation.Field(&b.MsgAckDatas, validation.Required),
	)

	return helpers.ValidationError(err, "user_bodyValidators.go", "groupChatMsgsAckDataT")

}

// ---------

type changeProfilePictureBody struct {
	PictureData []byte `json:"pictureData"`
}

func (b changeProfilePictureBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.PictureData,
			validation.Required,
			validation.Length(1, 2*1024*1024).Error("maximum picture size of 2mb exceeded"),
		),
	)

	return helpers.ValidationError(err, "user_bodyValidators.go", "changeProfilePictureBody")

}

type findNearbyUsersBody struct {
	LiveLocation *appTypes.UserGeolocation `json:"liveLocation"`
	Radius       float64                   `json:"radius"`
}

func (b findNearbyUsersBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.LiveLocation, validation.Required),
		validation.Field(&b.Radius, is.Float.Error("value must be of type float")),
	)

	return helpers.ValidationError(err, "user_bodyValidators.go", "findNearbyUsersBody")

}

type updateMyGeolocationBody struct {
	NewGeolocation string `json:"newGeolocation"`
}

func (b updateMyGeolocationBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.NewGeolocation,
			validation.Required,
			validation.Match(regexp.MustCompile("^[0-9]+, ?[0-9]+, ?[0-9]+$")).Error("invalid circle format; format: pointX, pointY, radiusR"),
		),
	)

	return helpers.ValidationError(err, "user_bodyValidators.go", "updateMyGeolocationBody")

}
