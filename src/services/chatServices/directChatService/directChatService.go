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
	})

	go func(msgData directChat.NewMessage) {
		var err error

		msgData.Sender, err = cache.GetUser[UITypes.ClientUser](context.Background(), clientUsername)
		if err != nil {
			return
		}
		err = cloudStorageService.MessageMediaCloudNameToUrl(msgData.Content)
		if err != nil {
			return
		}

		realtimeService.SendEventMsg(partnerUsername, appTypes.ServerEventMsg{
			Event: "direct chat: new message",
			Data:  msgData,
		})
	}(newMessage)

	return map[string]any{"new_msg_id": newMessage.Id}, nil
}

func AckMessageDelivered(ctx context.Context, clientUsername, partnerUsername, msgId string, deliveredAt int64) (any, error) {
	done, err := directChat.AckMessageDelivered(ctx, clientUsername, partnerUsername, msgId, deliveredAt)
	if err != nil {
		return nil, err
	}

	if done {
		// queue msg ack event
		go eventStreamService.QueueDirectMsgAckEvent(eventTypes.DirectMsgAckEvent{
			FromUser: clientUsername,
			ToUser:   partnerUsername,
			CHEId:    msgId,
			Ack:      "delivered",
			At:       deliveredAt,
		})

		go realtimeService.SendEventMsg(partnerUsername, appTypes.ServerEventMsg{
			Event: "direct chat: message delivered",
			Data: map[string]any{
				"partner_username": clientUsername,
				"msg_id":           msgId,
			},
		})
	}

	return done, nil
}

func AckMessageRead(ctx context.Context, clientUsername, partnerUsername, msgId string, readAt int64) (any, error) {
	done, err := directChat.AckMessageRead(ctx, clientUsername, partnerUsername, msgId, readAt)
	if err != nil {
		return nil, err
	}

	if done {
		// queue msg ack event
		go eventStreamService.QueueDirectMsgAckEvent(eventTypes.DirectMsgAckEvent{
			FromUser: clientUsername,
			ToUser:   partnerUsername,
			CHEId:    msgId,
			Ack:      "read",
			At:       readAt,
		})

		go realtimeService.SendEventMsg(partnerUsername, appTypes.ServerEventMsg{
			Event: "direct chat: message read",
			Data: map[string]any{
				"partner_username": clientUsername,
				"msg_id":           msgId,
			},
		})
	}

	return done, nil
}

func ReactToMessage(ctx context.Context, clientUsername, partnerUsername, msgId, emoji string, at int64) (any, error) {
	rxnToMessage, err := directChat.ReactToMessage(ctx, clientUsername, partnerUsername, msgId, emoji, at)
	if err != nil {
		return nil, err
	}

	done := rxnToMessage.CHEId != ""

	if !done {
		return done, nil
	}

	go func() {
		reactor, err := cache.GetUser[UITypes.ClientUser](context.Background(), clientUsername)
		if err != nil {
			return
		}

		realtimeService.SendEventMsg(partnerUsername, appTypes.ServerEventMsg{
			Event: "direct chat: message reaction",
			Data: map[string]any{
				"partner_username": clientUsername,
				"to_msg_id":        msgId,
				"reaction": UITypes.MsgReaction{
					Emoji:   emoji,
					Reactor: reactor,
				},
			},
		})
	}()

	// queue msg reaction event
	go eventStreamService.QueueNewDirectMsgReactionEvent(eventTypes.NewDirectMsgReactionEvent{
		FromUser: clientUsername,
		ToUser:   partnerUsername,
		CHEId:    rxnToMessage.CHEId,
		RxnData:  helpers.ToJson(rxnToMessage),
		ToMsgId:  msgId,
		Emoji:    emoji,
	})

	return done, nil
}

func RemoveReactionToMessage(ctx context.Context, clientUsername, partnerUsername, msgId string) (any, error) {
	CHEId, err := directChat.RemoveReactionToMessage(ctx, clientUsername, partnerUsername, msgId)
	if err != nil {
		return nil, err
	}

	done := CHEId != ""
	if !done {
		return done, nil
	}

	go realtimeService.SendEventMsg(partnerUsername, appTypes.ServerEventMsg{
		Event: "direct chat: message reaction removed",
		Data: map[string]any{
			"partner_username": clientUsername,
			"msg_id":           msgId,
		},
	})

	// queue reaction removed event
	go eventStreamService.QueueDirectMsgReactionRemovedEvent(eventTypes.DirectMsgReactionRemovedEvent{
		FromUser: clientUsername,
		ToUser:   partnerUsername,
		ToMsgId:  msgId,
		CHEId:    CHEId,
	})

	return done, nil
}

func GetChatHistory(ctx context.Context, clientUsername, partnerUsername string, limit int, cursor float64) (any, error) {
	return directChat.ChatHistory(ctx, clientUsername, partnerUsername, limit, cursor)
}
