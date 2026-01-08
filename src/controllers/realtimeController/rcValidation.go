package realtimeController

import (
	"i9chat/src/helpers"

	validation "github.com/go-ozzo/ozzo-validation/v4"
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

	return helpers.ValidationError(err, "rcValidation.go", "rtActionBody")
}

type subToUserPresenceAcd struct {
	Usernames []string `json:"users"`
}

func (vb subToUserPresenceAcd) Validate() error {
	err := validation.ValidateStruct(&vb,
		validation.Field(&vb.Usernames, validation.Required, validation.Length(1, 0)),
	)

	return helpers.ValidationError(err, "rcValidation.go", "subToUserPresenceAcd")
}

type unsubFromUserPresenceAcd struct {
	Usernames []string `json:"users"`
}

func (vb unsubFromUserPresenceAcd) Validate() error {
	err := validation.ValidateStruct(&vb,
		validation.Field(&vb.Usernames, validation.Required, validation.Length(1, 0)),
	)

	return helpers.ValidationError(err, "rcValidation.go", "unsubFromUserPresenceAcd")
}
