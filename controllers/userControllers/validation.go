package userControllers

import (
	"i9chat/appTypes"
	"i9chat/helpers"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type clientEventBody struct {
	Event string         `json:"event"`
	Data  map[string]any `json:"data"`
}

func (b clientEventBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.Event,
			validation.Required,
			validation.In(
				"new dm chat message",
				"dm chat message delivered ack",
				"dm chat message read ack",
				"new group chat message",
				"group chat message delivered ack",
				"group chat message read ack",
			).Error("invalid event"),
		),
		validation.Field(&b.Data, validation.Required),
	)

	return helpers.ValidationError(err, "user_bodyValidators.go", "clientEventBody")
}

type newDMChatMsg struct {
	PartnerUsername string               `json:"partnerUsername"`
	Msg             *appTypes.MsgContent `json:"msg"`
	CreatedAt       time.Time            `json:"at"`
}

func (vb newDMChatMsg) Validate() error {

	err := validation.ValidateStruct(&vb,
		validation.Field(&vb.PartnerUsername, validation.Required),
		validation.Field(&vb.Msg, validation.Required),
		validation.Field(&vb.CreatedAt,
			validation.Required,
			validation.Max(time.Now()).Error("invalid future time"),
		),
	)

	return helpers.ValidationError(err, "dmChat_bodyValidators.go", "newDMChatMsg")
}

type dmChatMsgAck struct {
	MsgId           string    `json:"msgId"`
	PartnerUsername string    `json:"partnerUsername"`
	At              time.Time `json:"at"`
}

func (d dmChatMsgAck) Validate() error {
	return validation.ValidateStruct(&d,
		validation.Field(&d.MsgId, validation.Required, is.UUID.Error("invalid id string")),
		validation.Field(&d.PartnerUsername, validation.Required),
		validation.Field(&d.At, validation.Required, validation.Max(time.Now()).Error("invalid future time")),
	)
}

type groupChatMsgAck struct {
	GroupChatId string    `json:"groupChatId"`
	MsgId       string    `json:"msgId"`
	At          time.Time `json:"at"`
}

func (d groupChatMsgAck) Validate() error {
	return validation.ValidateStruct(&d,
		validation.Field(&d.GroupChatId, validation.Required, is.UUID.Error("invalid id string")),
		validation.Field(&d.MsgId, validation.Required, is.UUID.Error("invalid id string")),
		validation.Field(&d.At, validation.Required, validation.Max(time.Now()).Error("invalid future time")),
	)
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

type findNearbyUsersQuery struct {
	Long   float64 `json:"long"`
	Lat    float64 `json:"lat"`
	Radius float64 `json:"radius"`
}

func (b findNearbyUsersQuery) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.Long, is.Float.Error("value must be of type float")),
		validation.Field(&b.Lat, is.Float.Error("value must be of type float")),
		validation.Field(&b.Radius, is.Float.Error("value must be of type float")),
	)

	return helpers.ValidationError(err, "user_bodyValidators.go", "findNearbyUsersQuery")

}

type updateMyGeolocationBody struct {
	NewGeolocation *appTypes.UserGeolocation `json:"newGeolocation"`
}

func (b updateMyGeolocationBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.NewGeolocation, validation.Required),
	)

	return helpers.ValidationError(err, "user_bodyValidators.go", "updateMyGeolocationBody")

}
