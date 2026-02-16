package directChatControllers

import (
	"context"
	"i9chat/src/appTypes"
	"i9chat/src/helpers"
	"i9chat/src/services/chatServices/directChatService"

	"github.com/goccy/go-json"
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

func SendMessage(ctx context.Context, clientUsername string, actionData json.RawMessage) (map[string]any, error) {

	acd := helpers.FromBtJson[sendDirectChatMsg](actionData)

	if err := acd.Validate(ctx); err != nil {
		return nil, err
	}

	return directChatService.SendMessage(ctx, clientUsername, acd.PartnerUsername, acd.ReplyTargetMsgId, acd.IsReply, helpers.ToJson(acd.Msg), acd.At)
}

func AckMessagesDelivered(ctx context.Context, clientUsername string, actionData json.RawMessage) (any, error) {

	acd := helpers.FromBtJson[directChatMsgsAck](actionData)

	if err := acd.Validate(); err != nil {
		return nil, err
	}

	return directChatService.AckMessagesDelivered(ctx, clientUsername, acd.PartnerUsername, acd.MsgIds, acd.At)
}

func AckMessagesRead(ctx context.Context, clientUsername string, actionData json.RawMessage) (any, error) {

	acd := helpers.FromBtJson[directChatMsgsAck](actionData)

	if err := acd.Validate(); err != nil {
		return nil, err
	}

	return directChatService.AckMessagesRead(ctx, clientUsername, acd.PartnerUsername, acd.MsgIds, acd.At)
}
