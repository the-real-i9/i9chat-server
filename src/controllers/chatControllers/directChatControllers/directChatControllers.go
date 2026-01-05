package directChatControllers

import (
	"i9chat/src/appTypes"
	"i9chat/src/services/chatServices/directChatService"

	"github.com/gofiber/fiber/v2"
)

func GetDirectChatHistory(c *fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, err := directChatService.GetChatHistory(ctx, clientUser.Username, c.Params("partner_username"), c.QueryInt("limit", 50), c.QueryFloat("cursor"))
	if err != nil {
		return err
	}

	return c.JSON(respData)
}
