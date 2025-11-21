package directChatService

import (
	"context"
	"i9chat/src/appTypes"
	"i9chat/src/appTypes/UITypes"
	"i9chat/src/helpers"
	directChat "i9chat/src/models/chatModel/directChatModel"
	"i9chat/src/services/appServices"
	"i9chat/src/services/realtimeService"
)

func GetChatHistory(ctx context.Context, clientUsername, partnerUsername string, limit int, offset int64) (any, error) {
	return directChat.ChatHistory(ctx, clientUsername, partnerUsername, limit, offset)
}

func SendMessage(ctx context.Context, clientUsername, partnerUsername, replyTargetMsgId string, isReply bool, msgContent *appTypes.MsgContent, at int64) (map[string]any, error) {

	err := appServices.UploadMessageMedia(ctx, clientUsername, msgContent)
	if err != nil {
		return nil, err
	}

	msgContentJson := helpers.ToJson(*msgContent)

	var newMessage directChat.NewMessageT

	if !isReply {
		newMessage, err = directChat.SendMessage(ctx, clientUsername, partnerUsername, msgContentJson, at)
		if err != nil {
			return nil, err
		}
	} else {
		newMessage, err = directChat.ReplyToMessage(ctx, clientUsername, partnerUsername, replyTargetMsgId, msgContentJson, at)
		if err != nil {
			return nil, err
		}
	}

	if newMessage.Id != "" {
		go func(msgData directChat.NewMessageT) {
			realtimeService.SendEventMsg(partnerUsername, appTypes.ServerEventMsg{
				Event: "direct chat: new message",
				Data:  msgData,
			})
		}(newMessage)

		// queue new message event
	}

	return map[string]any{"new_msg_id": newMessage.Id}, nil
}

func AckMessageDelivered(ctx context.Context, clientUsername, partnerUsername, msgId string, deliveredAt int64) (any, error) {
	done, err := directChat.AckMessageDelivered(ctx, clientUsername, partnerUsername, msgId, deliveredAt)
	if err != nil {
		return nil, err
	}

	if done {
		go realtimeService.SendEventMsg(partnerUsername, appTypes.ServerEventMsg{
			Event: "direct chat: message delivered",
			Data: map[string]any{
				"partner_username": clientUsername,
				"msg_id":           msgId,
				"delivered_at":     deliveredAt,
			},
		})

		// queue msg ack event
	}

	return done, nil
}

func AckMessageRead(ctx context.Context, clientUsername, partnerUsername, msgId string, readAt int64) (any, error) {
	done, err := directChat.AckMessageRead(ctx, clientUsername, partnerUsername, msgId, readAt)
	if err != nil {
		return nil, err
	}

	if done {
		go realtimeService.SendEventMsg(partnerUsername, appTypes.ServerEventMsg{
			Event: "direct chat: message read",
			Data: map[string]any{
				"partner_username": clientUsername,
				"msg_id":           msgId,
				"read_at":          readAt,
			},
		})

		// queue msg ack event
	}

	return done, nil
}

func ReactToMessage(ctx context.Context, clientUsername, partnerUsername, msgId, emoji string, at int64) (any, error) {
	rxnToMessage, err := directChat.ReactToMessage(ctx, clientUsername, partnerUsername, msgId, emoji, at)
	if err != nil {
		return nil, err
	}

	done := rxnToMessage.CHEId != ""

	if done {
		go realtimeService.SendEventMsg(partnerUsername, appTypes.ServerEventMsg{
			Event: "direct chat: message reaction",
			Data: map[string]any{
				"partner_username": clientUsername,
				"to_msg_id":        msgId,
				"reaction": UITypes.MsgReaction{
					Emoji:   emoji,
					Reactor: rxnToMessage.Reactor,
				},
			},
		})

		// queue msg reaction event
	}

	return done, nil
}

func RemoveReactionToMessage(ctx context.Context, clientUsername, partnerUsername, msgId string) (any, error) {
	CHEId, err := directChat.RemoveReactionToMessage(ctx, clientUsername, partnerUsername, msgId)
	if err != nil {
		return nil, err
	}

	done := CHEId != ""

	if done {
		go realtimeService.SendEventMsg(partnerUsername, appTypes.ServerEventMsg{
			Event: "direct chat: message reaction removed",
			Data: map[string]any{
				"partner_username": clientUsername,
				"msg_id":           msgId,
			},
		})
	}

	return done, nil
}
