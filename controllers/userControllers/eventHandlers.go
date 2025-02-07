package userControllers

import (
	"context"
	"i9chat/helpers"
	"i9chat/services/chatServices/dmChatService"
)

func newDMChatMsgEventHandler(ctx context.Context, clientUsername string, eventData map[string]any) (map[string]any, error) {
	var body newDMChatMsg

	helpers.MapToStruct(eventData, &body)

	if val_err := body.Validate(); val_err != nil {
		return nil, val_err
	}

	return dmChatService.SendMessage(ctx, clientUsername, body.PartnerUsername, body.Msg, body.CreatedAt)
}

func dmChatMsgDeliveredAckEventHandler(ctx context.Context, clientUsername string, eventData map[string]any) error {
	var body dmChatMsgAck

	helpers.MapToStruct(eventData, &body)

	if val_err := body.Validate(); val_err != nil {
		return val_err
	}

	return dmChatService.AckMessageDelivered(ctx, clientUsername, body.PartnerUsername, body.MsgId, body.At)
}

func dmChatMsgReadAckEventHandler(ctx context.Context, clientUsername string, eventData map[string]any) error {
	var body dmChatMsgAck

	helpers.MapToStruct(eventData, &body)

	if val_err := body.Validate(); val_err != nil {
		return val_err
	}

	return dmChatService.AckMessageRead(ctx, clientUsername, body.PartnerUsername, body.MsgId, body.At)
}
