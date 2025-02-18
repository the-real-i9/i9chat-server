package helpers

import (
	"fmt"
	"log"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/gofiber/fiber/v2"
)

func ValidationError(err error, filename, structname string) error {
	if err != nil {
		if e, ok := err.(validation.InternalError); ok {
			log.Printf("%s: %s: %v", filename, structname, e.InternalError())
			return fiber.ErrInternalServerError
		}

		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprint("validation error:", err))
	}

	return nil
}
