package realtimeController

import (
	"context"
	"i9chat/src/appTypes"
	"i9chat/src/helpers"
	"i9chat/src/services/chatServices/directChatService"
	"i9chat/src/services/chatServices/groupChatService"
)

func sendDirectChatMsgHndl(ctx context.Context, clientUser appTypes.ClientUser, actionData map[string]any) (map[string]any, error) {
	var acd sendDirectChatMsg

	helpers.ToStruct(actionData, &acd)

	if val_err := acd.Validate(); val_err != nil {
		return nil, val_err
	}

	return directChatService.SendMessage(ctx, clientUser, acd.PartnerUsername, acd.ReplyTargetMsgId, acd.IsReply, acd.Msg, acd.At)
}

func ackDirectChatMsgDeliveredHndl(ctx context.Context, clientUsername string, actionData map[string]any) (any, error) {
	var acd directChatMsgAck

	helpers.ToStruct(actionData, &acd)

	if val_err := acd.Validate(); val_err != nil {
		return nil, val_err
	}

	return directChatService.AckMessageDelivered(ctx, clientUsername, acd.PartnerUsername, acd.MsgId, acd.At)
}

func ackDirectChatMsgReadHndl(ctx context.Context, clientUsername string, actionData map[string]any) (any, error) {
	var acd directChatMsgAck

	helpers.ToStruct(actionData, &acd)

	if val_err := acd.Validate(); val_err != nil {
		return nil, val_err
	}

	return directChatService.AckMessageRead(ctx, clientUsername, acd.PartnerUsername, acd.MsgId, acd.At)
}

func sendGroupChatMsgHndl(ctx context.Context, clientUsername appTypes.ClientUser, actionData map[string]any) (map[string]any, error) {
	var acd sendGroupChatMsg

	helpers.ToStruct(actionData, &acd)

	if val_err := acd.Validate(); val_err != nil {
		return nil, val_err
	}

	return groupChatService.SendMessage(ctx, clientUsername, acd.GroupId, acd.ReplyTargetMsgId, acd.IsReply, acd.Msg, acd.At)
}

func ackGroupChatMsgDeliveredHndl(ctx context.Context, clientUsername string, actionData map[string]any) (any, error) {
	var acd groupChatMsgAck

	helpers.ToStruct(actionData, &acd)

	if val_err := acd.Validate(); val_err != nil {
		return nil, val_err
	}

	return groupChatService.AckMessageDelivered(ctx, clientUsername, acd.GroupId, acd.MsgId, acd.At)
}

func ackGroupChatMsgReadHndl(ctx context.Context, clientUsername string, actionData map[string]any) (any, error) {
	var acd groupChatMsgAck

	helpers.ToStruct(actionData, &acd)

	if val_err := acd.Validate(); val_err != nil {
		return nil, val_err
	}

	return groupChatService.AckMessageRead(ctx, clientUsername, acd.GroupId, acd.MsgId, acd.At)
}

func getGroupInfoHndl(ctx context.Context, actionData map[string]any) (map[string]any, error) {
	var acd groupInfo

	helpers.ToStruct(actionData, &acd)

	if val_err := acd.Validate(); val_err != nil {
		return nil, val_err
	}

	return groupChatService.GetGroupInfo(ctx, acd.GroupId)
}
