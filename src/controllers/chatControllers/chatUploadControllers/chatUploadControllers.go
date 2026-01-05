package chatUploadControllers

import (
	"i9chat/src/services/chatServices/chatUploadService"

	"github.com/gofiber/fiber/v2"
)

func AuthorizeUpload(c *fiber.Ctx) error {
	ctx := c.Context()

	var body authorizeUploadBody

	err := c.BodyParser(&body)
	if err != nil {
		return err
	}

	if err = body.Validate(); err != nil {
		return err
	}

	respData, err := chatUploadService.Authorize(ctx, body.MsgType, body.MediaMIME, body.MediaSize)
	if err != nil {
		return err
	}

	return c.JSON(respData)
}

func AuthorizeVisualUpload(c *fiber.Ctx) error {
	ctx := c.Context()

	var body authorizeVisualUploadBody

	err := c.BodyParser(&body)
	if err != nil {
		return err
	}

	if err = body.Validate(); err != nil {
		return err
	}

	respData, err := chatUploadService.AuthorizeVisual(ctx, body.MsgType, body.MediaMIME, body.MediaSize)
	if err != nil {
		return err
	}

	return c.JSON(respData)
}
