package realtimeController

import (
	"context"
	"i9chat/src/appTypes"
	"i9chat/src/helpers"
	"i9chat/src/services/chatServices/directChatService"
	"i9chat/src/services/chatServices/groupChatService"
)

func sendDirectChatMsgHndl(ctx context.Context, clientUser appTypes.ClientUser, actionData map[string]any) (map[string]any, error) {

	acd := helpers.ToStruct[sendDirectChatMsg](actionData)

	if err := acd.Validate(ctx); err != nil {
		return nil, err
	}

	return directChatService.SendMessage(ctx, clientUser, acd.PartnerUsername, acd.ReplyTargetMsgId, acd.IsReply, helpers.ToJson(acd.Msg), acd.At)
}

func ackDirectChatMsgDeliveredHndl(ctx context.Context, clientUsername string, actionData map[string]any) (any, error) {

	acd := helpers.ToStruct[directChatMsgAck](actionData)

	if err := acd.Validate(); err != nil {
		return nil, err
	}

	return directChatService.AckMessageDelivered(ctx, clientUsername, acd.PartnerUsername, acd.MsgId, acd.At)
}

func ackDirectChatMsgReadHndl(ctx context.Context, clientUsername string, actionData map[string]any) (any, error) {

	acd := helpers.ToStruct[directChatMsgAck](actionData)

	if err := acd.Validate(); err != nil {
		return nil, err
	}

	return directChatService.AckMessageRead(ctx, clientUsername, acd.PartnerUsername, acd.MsgId, acd.At)
}

func sendGroupChatMsgHndl(ctx context.Context, clientUsername appTypes.ClientUser, actionData map[string]any) (map[string]any, error) {

	acd := helpers.ToStruct[sendGroupChatMsg](actionData)

	if err := acd.Validate(ctx); err != nil {
		return nil, err
	}

	return groupChatService.SendMessage(ctx, clientUsername, acd.GroupId, acd.ReplyTargetMsgId, acd.IsReply, helpers.ToJson(acd.Msg), acd.At)
}

func ackGroupChatMsgDeliveredHndl(ctx context.Context, clientUsername string, actionData map[string]any) (any, error) {

	acd := helpers.ToStruct[groupChatMsgAck](actionData)

	if err := acd.Validate(); err != nil {
		return nil, err
	}

	return groupChatService.AckMessageDelivered(ctx, clientUsername, acd.GroupId, acd.MsgId, acd.At)
}

func ackGroupChatMsgReadHndl(ctx context.Context, clientUsername string, actionData map[string]any) (any, error) {

	acd := helpers.ToStruct[groupChatMsgAck](actionData)

	if err := acd.Validate(); err != nil {
		return nil, err
	}

	return groupChatService.AckMessageRead(ctx, clientUsername, acd.GroupId, acd.MsgId, acd.At)
}

func getGroupInfoHndl(ctx context.Context, actionData map[string]any) (any, error) {

	acd := helpers.ToStruct[groupInfo](actionData)

	if err := acd.Validate(); err != nil {
		return nil, err
	}

	return groupChatService.GetGroupInfo(ctx, acd.GroupId)
}
