package directChatControllers

import (
	"context"
	"i9chat/src/appTypes"
	"i9chat/src/helpers"
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

func SendMessage(ctx context.Context, clientUsername string, actionData map[string]any) (map[string]any, error) {

	acd := helpers.ToStruct[sendDirectChatMsg](actionData)

	if err := acd.Validate(ctx); err != nil {
		return nil, err
	}

	return directChatService.SendMessage(ctx, clientUsername, acd.PartnerUsername, acd.ReplyTargetMsgId, acd.IsReply, helpers.ToJson(acd.Msg), acd.At)
}

func AckMessageDelivered(ctx context.Context, clientUsername string, actionData map[string]any) (any, error) {

	acd := helpers.ToStruct[directChatMsgAck](actionData)

	if err := acd.Validate(); err != nil {
		return nil, err
	}

	return directChatService.AckMessageDelivered(ctx, clientUsername, acd.PartnerUsername, acd.MsgId, acd.At)
}

func AckMessageRead(ctx context.Context, clientUsername string, actionData map[string]any) (any, error) {

	acd := helpers.ToStruct[directChatMsgAck](actionData)

	if err := acd.Validate(); err != nil {
		return nil, err
	}

	return directChatService.AckMessageRead(ctx, clientUsername, acd.PartnerUsername, acd.MsgId, acd.At)
}
