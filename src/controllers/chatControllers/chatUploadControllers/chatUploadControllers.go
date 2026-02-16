package chatUploadControllers

import (
	"i9chat/src/services/chatServices/chatUploadService"

	"github.com/gofiber/fiber/v3"
)

func AuthorizeUpload(c fiber.Ctx) error {
	ctx := c.Context()

	var body authorizeUploadBody

	err := c.Bind().MsgPack(&body)
	if err != nil {
		return err
	}

	if err = body.Validate(); err != nil {
		return err
	}

	respData, err := chatUploadService.Authorize(ctx, body.MsgType, body.MediaMIME)
	if err != nil {
		return err
	}

	return c.MsgPack(respData)
}

func AuthorizeVisualUpload(c fiber.Ctx) error {
	ctx := c.Context()

	var body authorizeVisualUploadBody

	err := c.Bind().MsgPack(&body)
	if err != nil {
		return err
	}

	if err = body.Validate(); err != nil {
		return err
	}

	respData, err := chatUploadService.AuthorizeVisual(ctx, body.MsgType, body.MediaMIME)
	if err != nil {
		return err
	}

	return c.MsgPack(respData)
}
