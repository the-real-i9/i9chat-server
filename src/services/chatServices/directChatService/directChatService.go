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

	// queue new message event
	go eventStreamService.QueueNewDirectMessageEvent(eventTypes.NewDirectMessageEvent{
		FirstFromUser: newMessage.FirstFromUser,
		FirstToUser:   newMessage.FirstToUser,
		FromUser:      clientUsername,
		ToUser:        partnerUsername,
		CHEId:         newMessage.Id,
		MsgData:       helpers.ToJson(newMessage),
		CHECursor:     newMessage.Cursor,
	})

	go func(msg directChat.NewMessage) {
		uisender, _ := cache.GetUser[UITypes.ClientUser](context.Background(), clientUsername)

		uisender.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(uisender.ProfilePicUrl)

		cloudStorageService.MessageMediaCloudNameToUrl(msg.Content)

		UImsg := UITypes.ChatHistoryEntry{CHEType: msg.CHEType, Id: msg.Id, Content: msg.Content, DeliveryStatus: msg.DeliveryStatus, CreatedAt: msg.CreatedAt, Sender: uisender, ReplyTargetMsg: msg.ReplyTargetMsg, Cursor: msg.Cursor}

		if newMessage.FirstToUser {
			realtimeService.SendEventMsg(partnerUsername, appTypes.ServerEventMsg{
				Event: "new direct chat",
				Data: map[string]any{
					"chat":    UITypes.ChatSnippet{Type: "direct", PartnerUser: uisender, UnreadMC: 1, Cursor: msg.Cursor},
					"history": []UITypes.ChatHistoryEntry{UImsg},
				},
			})

			return
		}

		realtimeService.SendEventMsg(partnerUsername, appTypes.ServerEventMsg{
			Event: "direct chat: new che: message",
			Data:  UImsg,
		})
	}(newMessage)

	return map[string]any{"new_msg_id": newMessage.Id, "che_cursor": newMessage.Cursor}, nil
}

func AckMessageDelivered(ctx context.Context, clientUsername, partnerUsername, msgId string, deliveredAt int64) (map[string]any, error) {
	msgCursor, err := directChat.AckMessageDelivered(ctx, clientUsername, partnerUsername, msgId, deliveredAt)
	if err != nil {
		return nil, err
	}

	if msgCursor == nil {
		return nil, nil
	}

	// change this to the message serial number
	go realtimeService.SendEventMsg(partnerUsername, appTypes.ServerEventMsg{
		Event: "direct chat: message delivered",
		Data: map[string]any{
			"chat_partner": clientUsername,
			"msg_id":       msgId,
		},
	})

	// queue msg ack event
	go eventStreamService.QueueDirectMsgAckEvent(eventTypes.DirectMsgAckEvent{
		FromUser:   clientUsername,
		ToUser:     partnerUsername,
		CHEId:      msgId,
		Ack:        "delivered",
		At:         deliveredAt,
		ChatCursor: *msgCursor,
	})

	return map[string]any{"updated_chat_cursor": *msgCursor}, nil
}

func AckMessageRead(ctx context.Context, clientUsername, partnerUsername, msgId string, readAt int64) (bool, error) {
	done, err := directChat.AckMessageRead(ctx, clientUsername, partnerUsername, msgId, readAt)
	if err != nil {
		return true, err
	}

	if done {
		go realtimeService.SendEventMsg(partnerUsername, appTypes.ServerEventMsg{
			Event: "direct chat: message read",
			Data: map[string]any{
				"chat_partner": clientUsername,
				"msg_id":       msgId,
			},
		})

		// queue msg ack event
		go eventStreamService.QueueDirectMsgAckEvent(eventTypes.DirectMsgAckEvent{
			FromUser: clientUsername,
			ToUser:   partnerUsername,
			CHEId:    msgId,
			Ack:      "read",
			At:       readAt,
		})
	}

	return done, nil
}

func ReactToMessage(ctx context.Context, clientUsername, partnerUsername, msgId, emoji string, at int64) (map[string]any, error) {
	rxnToMessage, err := directChat.ReactToMessage(ctx, clientUsername, partnerUsername, msgId, emoji, at)
	if err != nil {
		return nil, err
	}

	if rxnToMessage.CHEId == "" {
		return nil, nil
	}

	go func(rxnData directChat.RxnToMessage) {
		uireactor, _ := cache.GetUser[UITypes.MsgReactor](context.Background(), clientUsername)

		uireactor.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(uireactor.ProfilePicUrl)

		realtimeService.SendEventMsg(partnerUsername, appTypes.ServerEventMsg{
			Event: "direct chat: message reaction",
			Data: map[string]any{
				"chat_partner": clientUsername,
				"che":          UITypes.ChatHistoryEntry{CHEType: rxnData.CHEType, Reactor: clientUsername, Emoji: rxnData.Emoji, Cursor: rxnData.Cursor},
				"msg_reaction": map[string]any{
					"msg_id": msgId,
					"reaction": UITypes.MsgReaction{
						Emoji:   emoji,
						Reactor: uireactor,
					},
				},
			},
		})
	}(rxnToMessage)

	// queue msg reaction event
	go func(rxnData directChat.RxnToMessage) {
		eventStreamService.QueueNewDirectMsgReactionEvent(eventTypes.NewDirectMsgReactionEvent{
			FromUser:  clientUsername,
			ToUser:    partnerUsername,
			CHEId:     rxnData.CHEId,
			RxnData:   helpers.ToJson(rxnData),
			ToMsgId:   msgId,
			Emoji:     emoji,
			CHECursor: rxnData.Cursor,
		})
	}(rxnToMessage)

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

func GetChatHistory(ctx context.Context, clientUsername, partnerUsername string, limit int, cursor float64) (any, error) {
	return directChat.ChatHistory(ctx, clientUsername, partnerUsername, limit, cursor)
}
