package dmChatControllers

import (
	"i9chat/helpers"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type getChatHistoryQuery struct {
	Limit  int       `json:"limit"`
	Offset time.Time `json:"offset"`
}

func (b getChatHistoryQuery) Validate() error {

	err := validation.ValidateStruct(&b,
		validation.Field(&b.Limit, validation.Required, validation.Min(1).Error("invalid value")),
		validation.Field(&b.Offset, validation.Required, validation.Min(time.Now()).Error("invalid future time")),
	)

	return helpers.ValidationError(err, "dmChat_validation.go", "getChatHistoryQuery")
}
