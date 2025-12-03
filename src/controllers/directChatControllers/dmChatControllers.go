package directChatControllers

import (
	"i9chat/src/appTypes"
	"i9chat/src/services/chatServices/directChatService"

	"github.com/gofiber/fiber/v2"
)

func GetDirectChatHistory(c *fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, app_err := directChatService.GetChatHistory(ctx, clientUser.Username, c.Params("partner_username"), c.QueryInt("limit", 50), c.QueryFloat("cursor"))
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}
