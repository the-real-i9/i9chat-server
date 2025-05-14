package realtimeController

import (
	"context"
	"i9chat/src/helpers"
	"i9chat/src/services/chatServices/dmChatService"
	"i9chat/src/services/chatServices/groupChatService"
	"time"
)

func sendDMChatMsgHndl(ctx context.Context, clientUsername string, eventData map[string]any) (map[string]any, error) {
	var evd sendDMChatMsg

	helpers.ToStruct(eventData, &evd)

	if val_err := evd.Validate(); val_err != nil {
		return nil, val_err
	}

	return dmChatService.SendMessage(ctx, clientUsername, evd.PartnerUsername, evd.Msg, evd.At)
}

func ackDMChatMsgDeliveredHndl(ctx context.Context, clientUsername string, eventData map[string]any) error {
	var evd dmChatMsgAck

	helpers.ToStruct(eventData, &evd)

	if val_err := evd.Validate(); val_err != nil {
		return val_err
	}

	return dmChatService.AckMessageDelivered(ctx, clientUsername, evd.PartnerUsername, evd.MsgId, evd.At)
}

func ackDMChatMsgReadHndl(ctx context.Context, clientUsername string, eventData map[string]any) error {
	var evd dmChatMsgAck

	helpers.ToStruct(eventData, &evd)

	if val_err := evd.Validate(); val_err != nil {
		return val_err
	}

	return dmChatService.AckMessageRead(ctx, clientUsername, evd.PartnerUsername, evd.MsgId, evd.At)
}

func sendGroupChatMsgHndl(ctx context.Context, clientUsername string, eventData map[string]any) (map[string]any, error) {
	var evd sendGroupChatMsg

	helpers.ToStruct(eventData, &evd)

	if val_err := evd.Validate(); val_err != nil {
		return nil, val_err
	}

	return groupChatService.SendMessage(ctx, clientUsername, evd.GroupId, evd.Msg, evd.At)
}

func ackGroupChatMsgDeliveredHndl(ctx context.Context, clientUsername string, eventData map[string]any) (any, error) {
	var evd groupChatMsgAck

	helpers.ToStruct(eventData, &evd)

	if val_err := evd.Validate(); val_err != nil {
		return nil, val_err
	}

	return groupChatService.AckMessageDelivered(ctx, clientUsername, evd.GroupId, evd.MsgId, evd.At)
}

func ackGroupChatMsgReadHndl(ctx context.Context, clientUsername string, eventData map[string]any) (any, error) {
	var evd groupChatMsgAck

	helpers.ToStruct(eventData, &evd)

	if val_err := evd.Validate(); val_err != nil {
		return nil, val_err
	}

	return groupChatService.AckMessageRead(ctx, clientUsername, evd.GroupId, evd.MsgId, evd.At)
}

func getGroupInfoHndl(ctx context.Context, eventData map[string]any) (map[string]any, error) {
	var evd groupInfo

	helpers.ToStruct(eventData, &evd)

	if val_err := evd.Validate(); val_err != nil {
		return nil, val_err
	}

	return groupChatService.GetGroupInfo(ctx, evd.GroupId)
}

func getDMChatHistoryHndl(ctx context.Context, clientUsername string, eventData map[string]any) ([]any, error) {
	var evd dmChatHistory

	helpers.ToStruct(eventData, &evd)

	if val_err := evd.Validate(); val_err != nil {
		return nil, val_err
	}

	if evd.Offset == 0 {
		evd.Offset = time.Now().UTC().UnixMilli()
	}

	if evd.Limit == 0 {
		evd.Limit = 50
	}

	return dmChatService.GetChatHistory(ctx, clientUsername, evd.PartnerUsername, evd.Limit, evd.Offset)
}

func getGroupChatHistoryHndl(ctx context.Context, clientUsername string, eventData map[string]any) (any, error) {
	var evd groupChatHistory

	helpers.ToStruct(eventData, &evd)

	if val_err := evd.Validate(); val_err != nil {
		return nil, val_err
	}

	if evd.Offset == 0 {
		evd.Offset = time.Now().UTC().UnixMilli()
	}

	if evd.Limit == 0 {
		evd.Limit = 50
	}

	return groupChatService.GetChatHistory(ctx, clientUsername, evd.GroupId, evd.Limit, evd.Offset)
}
