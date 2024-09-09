package helpers

import (
	"i9chat/appTypes"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

var UserSliceRule validation.RuleFunc = func(value any) error {
	user := value.([]appTypes.String)

	if err := validation.Validate(user[0], validation.Required, is.Int.Error("invalid non-integer value")); err != nil {
		return err
	}

	if err := validation.Validate(user[1], validation.Required); err != nil {
		return err
	}

	return nil
}
