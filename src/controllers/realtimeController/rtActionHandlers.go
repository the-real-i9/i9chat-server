package realtimeController

import (
	"context"
	"i9chat/src/helpers"
	dmChat "i9chat/src/models/chatModel/dmChatModel"
	"i9chat/src/services/chatServices/dmChatService"
	"i9chat/src/services/chatServices/groupChatService"
	"time"
)

func sendDMChatMsgHndl(ctx context.Context, clientUsername string, actionData map[string]any) (map[string]any, error) {
	var acd sendDMChatMsg

	helpers.ToStruct(actionData, &acd)

	if val_err := acd.Validate(); val_err != nil {
		return nil, val_err
	}

	return dmChatService.SendMessage(ctx, clientUsername, acd.PartnerUsername, acd.Msg, acd.At)
}

func ackDMChatMsgDeliveredHndl(ctx context.Context, clientUsername string, actionData map[string]any) (any, error) {
	var acd dmChatMsgAck

	helpers.ToStruct(actionData, &acd)

	if val_err := acd.Validate(); val_err != nil {
		return nil, val_err
	}

	return dmChatService.AckMessageDelivered(ctx, clientUsername, acd.PartnerUsername, acd.MsgId, acd.At)
}

func ackDMChatMsgReadHndl(ctx context.Context, clientUsername string, actionData map[string]any) (any, error) {
	var acd dmChatMsgAck

	helpers.ToStruct(actionData, &acd)

	if val_err := acd.Validate(); val_err != nil {
		return nil, val_err
	}

	return dmChatService.AckMessageRead(ctx, clientUsername, acd.PartnerUsername, acd.MsgId, acd.At)
}

func sendGroupChatMsgHndl(ctx context.Context, clientUsername string, actionData map[string]any) (map[string]any, error) {
	var acd sendGroupChatMsg

	helpers.ToStruct(actionData, &acd)

	if val_err := acd.Validate(); val_err != nil {
		return nil, val_err
	}

	return groupChatService.SendMessage(ctx, clientUsername, acd.GroupId, acd.Msg, acd.At)
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

func getDMChatHistoryHndl(ctx context.Context, clientUsername string, actionData map[string]any) ([]dmChat.ChatHistoryEntry, error) {
	var acd dmChatHistory

	helpers.ToStruct(actionData, &acd)

	if val_err := acd.Validate(); val_err != nil {
		return nil, val_err
	}

	if acd.Offset == 0 {
		acd.Offset = time.Now().UTC().UnixMilli()
	}

	if acd.Limit == 0 {
		acd.Limit = 50
	}

	return dmChatService.GetChatHistory(ctx, clientUsername, acd.PartnerUsername, acd.Limit, acd.Offset)
}

func getGroupChatHistoryHndl(ctx context.Context, clientUsername string, actionData map[string]any) (any, error) {
	var acd groupChatHistory

	helpers.ToStruct(actionData, &acd)

	if val_err := acd.Validate(); val_err != nil {
		return nil, val_err
	}

	if acd.Offset == 0 {
		acd.Offset = time.Now().UTC().UnixMilli()
	}

	if acd.Limit == 0 {
		acd.Limit = 50
	}

	return groupChatService.GetChatHistory(ctx, clientUsername, acd.GroupId, acd.Limit, acd.Offset)
}
