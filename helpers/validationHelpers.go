package helpers

import (
	"i9chat/appGlobals"
	"i9chat/appTypes"
	"log"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/gofiber/fiber/v2"
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

func ValidationError(err error, filename, structname string) error {
	if err != nil {
		if e, ok := err.(validation.InternalError); ok {
			log.Printf("%s: %s: %v", filename, structname, e.InternalError())
			return appGlobals.ErrInternalServerError
		}

		return fiber.NewError(fiber.StatusBadRequest, "validation error:", err.Error())
	}

	return nil
}
