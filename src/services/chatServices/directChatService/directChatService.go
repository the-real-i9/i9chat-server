package directChatService

import (
	"context"
	"i9chat/src/appTypes"
	"i9chat/src/appTypes/UITypes"
	"i9chat/src/cache"
	"i9chat/src/helpers"
	directChat "i9chat/src/models/chatModel/directChatModel"
	"i9chat/src/services/cloudStorageService"
	"i9chat/src/services/eventStreamService"
	"i9chat/src/services/eventStreamService/eventTypes"
	"i9chat/src/services/realtimeService"
)

func SendMessage(ctx context.Context, clientUsername, partnerUsername, replyTargetMsgId string, isReply bool, msgContentJson string, at int64) (map[string]any, error) {
	var (
		newMessage directChat.NewMessage
		err        error
	)

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

	if newMessage.Id == "" {
		return nil, nil
	}

	go func(msg directChat.NewMessage, clientUsername, partnerUsername string) {
		uisender, _ := cache.GetUser[UITypes.ClientUser](context.Background(), clientUsername)

		uisender.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(uisender.ProfilePicUrl)

		UImsg := UITypes.ChatHistoryEntry{
			CHEType: msg.CHEType, Id: msg.Id,
			Content:        cloudStorageService.MessageMediaCloudNameToUrl(msg.Content),
			DeliveryStatus: msg.DeliveryStatus, CreatedAt: msg.CreatedAt, Sender: uisender,
			ReplyTargetMsg: msg.ReplyTargetMsg, Cursor: float64(msg.Cursor),
		}

		if newMessage.FirstToUser {
			realtimeService.SendEventMsg(partnerUsername, appTypes.ServerEventMsg{
				Event: "new direct chat",
				Data: map[string]any{
					"chat":    UITypes.ChatSnippet{Type: "direct", PartnerUser: uisender, UnreadMC: 1, Cursor: float64(msg.Cursor)},
					"history": []UITypes.ChatHistoryEntry{UImsg},
				},
			})

			return
		}

		realtimeService.SendEventMsg(partnerUsername, appTypes.ServerEventMsg{
			Event: "direct chat: new che: message",
			Data:  UImsg,
		})
	}(newMessage, clientUsername, partnerUsername)

	// queue new message event
	go func(newMessage directChat.NewMessage, clientUsername, partnerUsername string) {
		eventStreamService.QueueNewDirectMessageEvent(eventTypes.NewDirectMessageEvent{
			FirstFromUser: newMessage.FirstFromUser,
			FirstToUser:   newMessage.FirstToUser,
			FromUser:      clientUsername,
			ToUser:        partnerUsername,
			CHEId:         newMessage.Id,
			MsgData:       helpers.ToMsgPack(newMessage),
			CHECursor:     newMessage.Cursor,
		})
	}(newMessage, clientUsername, partnerUsername)

	return map[string]any{"new_msg_id": newMessage.Id, "che_cursor": newMessage.Cursor}, nil
}

func AckMessagesDelivered(ctx context.Context, clientUsername, partnerUsername string, msgIds []any, deliveredAt int64) (map[string]any, error) {
	lastMsgCursor, err := directChat.AckMessageDelivered(ctx, clientUsername, partnerUsername, msgIds, deliveredAt)
	if err != nil {
		return nil, err
	}

	if lastMsgCursor == 0 {
		return nil, nil
	}

	go realtimeService.SendEventMsg(partnerUsername, appTypes.ServerEventMsg{
		Event: "direct chat: messages delivered",
		Data: map[string]any{
			"chat_partner": clientUsername,
			"delivered_at": deliveredAt,
			"msg_ids":      msgIds,
		},
	})

	// queue msg ack event
	go eventStreamService.QueueDirectMsgAckEvent(eventTypes.DirectMsgAckEvent{
		FromUser:   clientUsername,
		ToUser:     partnerUsername,
		CHEIds:     msgIds,
		Ack:        "delivered",
		At:         deliveredAt,
		ChatCursor: lastMsgCursor,
	})

	return map[string]any{"updated_chat_cursor": lastMsgCursor}, nil
}

func AckMessagesRead(ctx context.Context, clientUsername, partnerUsername string, msgIds []any, readAt int64) (bool, error) {
	done, err := directChat.AckMessageRead(ctx, clientUsername, partnerUsername, msgIds, readAt)
	if err != nil {
		return true, err
	}

	if done {
		go realtimeService.SendEventMsg(partnerUsername, appTypes.ServerEventMsg{
			Event: "direct chat: messages read",
			Data: map[string]any{
				"chat_partner": clientUsername,
				"read_at":      readAt,
				"msg_ids":      msgIds,
			},
		})

		// queue msg ack event
		go eventStreamService.QueueDirectMsgAckEvent(eventTypes.DirectMsgAckEvent{
			FromUser: clientUsername,
			ToUser:   partnerUsername,
			CHEIds:   msgIds,
			Ack:      "read",
			At:       readAt,
		})
	}

	return done, nil
}

// Fix business logic: There's the possiblitity of the message not existing in the partner's user's chat
// Do the client's first, then return the condition whether the message exists with the partner
// If true, then do for the partner, else skip for them
func ReactToMessage(ctx context.Context, clientUsername, partnerUsername, msgId, emoji string, at int64) (map[string]any, error) {
	rxnToMessage, err := directChat.ReactToMessage(ctx, clientUsername, partnerUsername, msgId, emoji, at)
	if err != nil {
		return nil, err
	}

	if rxnToMessage.CHEId == "" {
		return nil, nil
	}

	go func(rxnData directChat.RxnToMessage, clientUsername, partnerUsername string) {
		uireactor, _ := cache.GetUser[UITypes.MsgReactor](context.Background(), clientUsername)

		uireactor.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(uireactor.ProfilePicUrl)

		realtimeService.SendEventMsg(partnerUsername, appTypes.ServerEventMsg{
			Event: "direct chat: message reaction",
			Data: map[string]any{
				"chat_partner": clientUsername,
				"che":          UITypes.ChatHistoryEntry{CHEType: rxnData.CHEType, Reactor: clientUsername, Emoji: rxnData.Emoji, Cursor: float64(rxnData.Cursor)},
				"msg_reaction": map[string]any{
					"msg_id": rxnData.ToMsgId,
					"reaction": UITypes.MsgReaction{
						Emoji:   rxnData.Emoji,
						Reactor: uireactor,
					},
				},
			},
		})
	}(rxnToMessage, clientUsername, partnerUsername)

	// queue msg reaction event
	go func(rxnData directChat.RxnToMessage, clientUsername, partnerUsername string) {
		eventStreamService.QueueNewDirectMsgReactionEvent(eventTypes.NewDirectMsgReactionEvent{
			FromUser:  clientUsername,
			ToUser:    partnerUsername,
			CHEId:     rxnData.CHEId,
			RxnData:   helpers.ToMsgPack(rxnData),
			ToMsgId:   rxnData.ToMsgId,
			Emoji:     rxnData.Emoji,
			CHECursor: rxnData.Cursor,
		})
	}(rxnToMessage, clientUsername, partnerUsername)

	return map[string]any{"che_cursor": rxnToMessage.Cursor}, nil
}

func RemoveReactionToMessage(ctx context.Context, clientUsername, partnerUsername, msgId string) (bool, error) {
	CHEId, err := directChat.RemoveReactionToMessage(ctx, clientUsername, partnerUsername, msgId)
	if err != nil {
		return false, err
	}

	done := CHEId != ""

	if done {
		go realtimeService.SendEventMsg(partnerUsername, appTypes.ServerEventMsg{
			Event: "direct chat: message reaction removed",
			Data: map[string]any{
				"chat_partner": clientUsername,
				"msg_id":       msgId,
			},
		})

		// queue reaction removed event
		go eventStreamService.QueueDirectMsgReactionRemovedEvent(eventTypes.DirectMsgReactionRemovedEvent{
			FromUser: clientUsername,
			ToUser:   partnerUsername,
			ToMsgId:  msgId,
			CHEId:    CHEId,
		})
	}

	return done, nil
}

func GetChatHistory(ctx context.Context, clientUsername, partnerUsername string, limit int64, cursor float64) (any, error) {
	return directChat.ChatHistory(ctx, clientUsername, partnerUsername, limit, cursor)
}
