package appModel

import (
	"context"
	"i9chat/helpers"
	"log"

	"github.com/gofiber/fiber/v2"
)

func AccountExists(ctx context.Context, emailOrUsername string) (bool, error) {
	exist, err := helpers.QueryRowField[bool](ctx, "SELECT exist FROM account_exists($1)", emailOrUsername)

	if err != nil {
		log.Println("appModel.go: AccountExists:", err)
		return false, fiber.ErrInternalServerError
	}

	return *exist, nil
}
