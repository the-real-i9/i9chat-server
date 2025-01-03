package dmChatControllers

import (
	"context"
	"fmt"
	"i9chat/appTypes"
	"i9chat/services/chatServices/dmChatService"
	"log"

	"github.com/gofiber/fiber/v2"
)

func SendMessage(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(*appTypes.ClientUser)

	var partnerUserId int

	_, err := fmt.Sscanf(c.Params("partner_user_id"), "%d", &partnerUserId)
	if err != nil {
		log.Println("DMChatControllers.go: SendMessage: fmt.Sscanf:", err)
		return fiber.ErrInternalServerError
	}

	var body sendMessageBody

	body_err := c.BodyParser(&body)
	if body_err != nil {
		return body_err
	}

	if val_err := body.Validate(); val_err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(val_err.Error())
	}

	respData, app_err := dmChatService.SendMessage(ctx,
		clientUser.Id,
		partnerUserId,
		body.Msg,
		body.At,
	)
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func GetChatHistory(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var body getChatHistoryBody

	body_err := c.BodyParser(&body)
	if body_err != nil {
		return body_err
	}

	if val_err := body.Validate(); val_err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(val_err.Error())
	}

	respData, app_err := dmChatService.GetChatHistory(ctx, body.DMChatId, body.Offset)
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}
