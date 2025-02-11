package dmChatControllers

import (
	"context"
	"i9chat/appTypes"
	"i9chat/services/chatServices/dmChatService"

	"github.com/gofiber/fiber/v2"
)

func GetChatHistory(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(*appTypes.ClientUser)

	partnerUsername := c.Params("partner_username")

	var query getChatHistoryQuery

	query_err := c.QueryParser(&query)
	if query_err != nil {
		return query_err
	}

	if val_err := query.Validate(); val_err != nil {
		return fiber.NewError(fiber.StatusBadRequest, val_err.Error())
	}

	respData, app_err := dmChatService.GetChatHistory(ctx, clientUser.Username, partnerUsername, query.Limit, query.Offset)
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}
