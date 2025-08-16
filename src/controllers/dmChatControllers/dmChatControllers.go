package dmChatControllers

import (
	"context"
	"i9chat/src/appTypes"
	"i9chat/src/services/chatServices/dmChatService"
	"time"

	"github.com/gofiber/fiber/v2"
)

func GetDMChatHistory(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, app_err := dmChatService.GetChatHistory(ctx, clientUser.Username, c.Params("partner_username"), c.QueryInt("limit", 50), int64(c.QueryInt("offset", int(time.Now().UTC().UnixMilli()))))
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}
