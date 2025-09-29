package dmChatService

import (
	"context"
	"encoding/json"
	"i9chat/src/appTypes"
	dmChat "i9chat/src/models/chatModel/dmChatModel"
	"i9chat/src/services/appServices"
	"i9chat/src/services/eventStreamService"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

func GetChatHistory(ctx context.Context, clientUsername, partnerUsername string, limit int, offset int64) (any, error) {
	return dmChat.ChatHistory(ctx, clientUsername, partnerUsername, limit, time.UnixMilli(offset).UTC())
}

func SendMessage(ctx context.Context, clientUsername, partnerUsername, replyTargetMsgId string, isReply bool, msgContent *appTypes.MsgContent, at int64) (map[string]any, error) {

	err := appServices.UploadMessageMedia(ctx, clientUsername, msgContent)
	if err != nil {
		return nil, err
	}

	msgContentJson, err := json.Marshal(*msgContent)
	if err != nil {
		log.Println("DMChatService.go: SendMessage: json.Marshal:", err)
		return nil, fiber.ErrInternalServerError
	}

	var newMessage dmChat.NewMessage

	if !isReply {
		newMessage, err = dmChat.SendMessage(ctx, clientUsername, partnerUsername, string(msgContentJson), time.UnixMilli(at).UTC())
		if err != nil {
			return nil, err
		}
	} else {
		newMessage, err = dmChat.ReplyToMessage(ctx, clientUsername, partnerUsername, replyTargetMsgId, string(msgContentJson), time.UnixMilli(at).UTC())
		if err != nil {
			return nil, err
		}
	}

	if newMessage.PartnerData != nil {
		go eventStreamService.Send(partnerUsername, appTypes.ServerWSMsg{
			Event: "new dm chat message",
			Data:  newMessage.PartnerData,
		})
	}

	return newMessage.ClientData, nil
}

func AckMessageDelivered(ctx context.Context, clientUsername, partnerUsername, msgId string, deliveredAt int64) (any, error) {
	done, err := dmChat.AckMessageDelivered(ctx, clientUsername, partnerUsername, msgId, time.UnixMilli(deliveredAt).UTC())
	if err != nil {
		return nil, err
	}

	if done {
		go eventStreamService.Send(partnerUsername, appTypes.ServerWSMsg{
			Event: "dm chat message delivered",
			Data: map[string]any{
				"partner_username": clientUsername,
				"msg_id":           msgId,
			},
		})
	}

	return done, nil
}

func AckMessageRead(ctx context.Context, clientUsername, partnerUsername, msgId string, readAt int64) (any, error) {
	done, err := dmChat.AckMessageRead(ctx, clientUsername, partnerUsername, msgId, time.UnixMilli(readAt).UTC())
	if err != nil {
		return nil, err
	}

	if done {
		go eventStreamService.Send(partnerUsername, appTypes.ServerWSMsg{
			Event: "dm chat message read",
			Data: map[string]any{
				"partner_username": clientUsername,
				"msg_id":           msgId,
			},
		})
	}

	return done, nil
}

func ReactToMessage(ctx context.Context, clientUsername, partnerUsername, msgId, reaction string, at int64) (any, error) {
	rxnToMessage, err := dmChat.ReactToMessage(ctx, clientUsername, partnerUsername, msgId, reaction, time.UnixMilli(at).UTC())
	if err != nil {
		return nil, err
	}

	if rxnToMessage.PartnerData != nil {
		go eventStreamService.Send(partnerUsername, appTypes.ServerWSMsg{
			Event: "dm chat message reaction",
			Data:  rxnToMessage.PartnerData,
		})
	}

	return rxnToMessage.ClientData, nil
}

func RemoveReactionToMessage(ctx context.Context, clientUsername, partnerUsername, msgId string) (any, error) {
	done, err := dmChat.RemoveReactionToMessage(ctx, clientUsername, partnerUsername, msgId)
	if err != nil {
		return nil, err
	}

	if done {
		go eventStreamService.Send(partnerUsername, appTypes.ServerWSMsg{
			Event: "dm chat message reaction removed",
			Data: map[string]any{
				"partner_username": clientUsername,
				"msg_id":           msgId,
				"reactor_username": clientUsername,
			},
		})
	}

	return done, nil
}
