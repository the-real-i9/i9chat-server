package directChatControllers

import (
	"context"
	"i9chat/src/appTypes"
	"i9chat/src/helpers"
	"i9chat/src/services/chatServices/directChatService"

	"github.com/gofiber/fiber/v3"
	"github.com/vmihailenco/msgpack/v5"
)

func GetDirectChatHistory(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var query struct {
		Limit  int64
		Cursor float64
	}

	if err := c.Bind().Query(&query); err != nil {
		return err
	}

	respData, err := directChatService.GetChatHistory(ctx, clientUser.Username, c.Params("partner_username"), helpers.CoalesceInt(query.Limit, 50), query.Cursor)
	if err != nil {
		return err
	}

	return c.MsgPack(respData)
}

func SendMessage(ctx context.Context, clientUsername string, actionData msgpack.RawMessage) (map[string]any, error) {

	acd := helpers.FromBtMsgPack[sendDirectChatMsg](actionData)

	if err := acd.Validate(ctx); err != nil {
		return nil, err
	}

	return directChatService.SendMessage(ctx, clientUsername, acd.PartnerUsername, acd.ReplyTargetMsgId, acd.IsReply, helpers.ToJson(acd.Msg), acd.At)
}

func AckMessagesDelivered(ctx context.Context, clientUsername string, actionData msgpack.RawMessage) (any, error) {

	acd := helpers.FromBtMsgPack[directChatMsgsAck](actionData)

	if err := acd.Validate(); err != nil {
		return nil, err
	}

	return directChatService.AckMessagesDelivered(ctx, clientUsername, acd.PartnerUsername, acd.MsgIds, acd.At)
}

func AckMessagesRead(ctx context.Context, clientUsername string, actionData msgpack.RawMessage) (any, error) {

	acd := helpers.FromBtMsgPack[directChatMsgsAck](actionData)

	if err := acd.Validate(); err != nil {
		return nil, err
	}

	return directChatService.AckMessagesRead(ctx, clientUsername, acd.PartnerUsername, acd.MsgIds, acd.At)
}
