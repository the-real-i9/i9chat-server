package userControllers

import (
	"context"
	"i9chat/helpers"
	"i9chat/services/chatServices/dmChatService"
	"i9chat/services/chatServices/groupChatService"
)

func newDMChatMsgEvHd(ctx context.Context, clientUsername string, eventData map[string]any) (map[string]any, error) {
	var body newDMChatMsg

	helpers.MapToStruct(eventData, &body)

	if val_err := body.Validate(); val_err != nil {
		return nil, val_err
	}

	return dmChatService.SendMessage(ctx, clientUsername, body.PartnerUsername, body.Msg, body.CreatedAt)
}

func dmChatMsgDeliveredAckEvHd(ctx context.Context, clientUsername string, eventData map[string]any) error {
	var body dmChatMsgAck

	helpers.MapToStruct(eventData, &body)

	if val_err := body.Validate(); val_err != nil {
		return val_err
	}

	return dmChatService.AckMessageDelivered(ctx, clientUsername, body.PartnerUsername, body.MsgId, body.At)
}

func dmChatMsgReadAckEvHd(ctx context.Context, clientUsername string, eventData map[string]any) error {
	var body dmChatMsgAck

	helpers.MapToStruct(eventData, &body)

	if val_err := body.Validate(); val_err != nil {
		return val_err
	}

	return dmChatService.AckMessageRead(ctx, clientUsername, body.PartnerUsername, body.MsgId, body.At)
}

func newGroupChatMsgEvHd(ctx context.Context, clientUsername string, eventData map[string]any) (map[string]any, error) {
	var body newGroupChatMsg

	helpers.MapToStruct(eventData, &body)

	if val_err := body.Validate(); val_err != nil {
		return nil, val_err
	}

	return groupChatService.SendMessage(ctx, clientUsername, body.GroupId, body.Msg, body.CreatedAt)
}

func groupChatMsgDeliveredAckEvHd(ctx context.Context, clientUsername string, eventData map[string]any) error {
	var body groupChatMsgAck

	helpers.MapToStruct(eventData, &body)

	if val_err := body.Validate(); val_err != nil {
		return val_err
	}

	return groupChatService.AckMessageDelivered(ctx, clientUsername, body.GroupId, body.MsgId, body.At)
}

func groupChatMsgReadAckEvHd(ctx context.Context, clientUsername string, eventData map[string]any) error {
	var body groupChatMsgAck

	helpers.MapToStruct(eventData, &body)

	if val_err := body.Validate(); val_err != nil {
		return val_err
	}

	return groupChatService.AckMessageRead(ctx, clientUsername, body.GroupId, body.MsgId, body.At)
}

func groupInfoEvHd(ctx context.Context, eventData map[string]any) (map[string]any, error) {
	var body groupInfo

	helpers.MapToStruct(eventData, &body)

	if val_err := body.Validate(); val_err != nil {
		return nil, val_err
	}

	return groupChatService.GetGroupInfo(ctx, body.GroupId)
}

func groupMemInfoEvHd(ctx context.Context, clientUsername string, eventData map[string]any) (map[string]any, error) {
	var body groupMemInfo

	helpers.MapToStruct(eventData, &body)

	if val_err := body.Validate(); val_err != nil {
		return nil, val_err
	}

	return groupChatService.GetGroupMemInfo(ctx, clientUsername, body.GroupId)
}
