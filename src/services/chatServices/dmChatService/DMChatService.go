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

func GetChatHistory(ctx context.Context, clientUsername, partnerUsername string, limit int, offset int64) ([]any, error) {
	return dmChat.ChatHistory(ctx, clientUsername, partnerUsername, limit, time.UnixMilli(offset).UTC())
}

func SendMessage(ctx context.Context, clientUsername, partnerUsername string, msgContent *appTypes.MsgContent, at int64) (map[string]any, error) {

	err := appServices.UploadMessageMedia(ctx, clientUsername, msgContent)
	if err != nil {
		return nil, err
	}

	msgContentJson, err := json.Marshal(*msgContent)
	if err != nil {
		log.Println("DMChatService.go: SendMessage: json.Marshal:", err)
		return nil, fiber.ErrInternalServerError
	}

	newMessage, err := dmChat.SendMessage(ctx, clientUsername, partnerUsername, string(msgContentJson), time.UnixMilli(at).UTC())
	if err != nil {
		return nil, err
	}

	go eventStreamService.Send(partnerUsername, appTypes.ServerWSMsg{
		Event: "new dm chat message",
		Data:  newMessage.PartnerData,
	})

	return newMessage.ClientData, nil
}

func AckMessageDelivered(ctx context.Context, clientUsername, partnerUsername, msgId string, deliveredAt int64) error {
	if err := dmChat.AckMessageDelivered(ctx, clientUsername, partnerUsername, msgId, time.UnixMilli(deliveredAt).UTC()); err != nil {
		return err
	}

	go eventStreamService.Send(partnerUsername, appTypes.ServerWSMsg{
		Event: "dm chat message delivered",
		Data: map[string]any{
			"partner_username": clientUsername,
			"msg_id":           msgId,
		},
	})
	return nil
}

func AckMessageRead(ctx context.Context, clientUsername, partnerUsername, msgId string, readAt int64) error {
	if err := dmChat.AckMessageRead(ctx, clientUsername, partnerUsername, msgId, time.UnixMilli(readAt).UTC()); err != nil {
		return err
	}
	go eventStreamService.Send(partnerUsername, appTypes.ServerWSMsg{
		Event: "dm chat message read",
		Data: map[string]any{
			"partner_username": clientUsername,
			"msg_id":           msgId,
		},
	})

	return nil
}

func React() {

}
