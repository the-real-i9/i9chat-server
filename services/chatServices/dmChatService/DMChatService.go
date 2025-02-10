package dmChatService

import (
	"context"
	"encoding/json"
	"fmt"
	"i9chat/appTypes"
	dmChat "i9chat/models/chatModel/dmChatModel"
	"i9chat/services/appServices"
	"i9chat/services/messageBrokerService"
	"time"
)

func GetChatHistory(ctx context.Context, clientUsername, partnerUsername string, limit int, offset time.Time) ([]any, error) {
	return dmChat.GetChatHistory(ctx, clientUsername, partnerUsername, limit, offset)
}

func SendMessage(ctx context.Context, clientUsername, partnerUsername string, msgContent *appTypes.MsgContent, createdAt time.Time) (map[string]any, error) {

	err := appServices.UploadMessageMedia(ctx, clientUsername, msgContent)
	if err != nil {
		return nil, err
	}

	msgContentJson, _ := json.Marshal(*msgContent)

	newMessage, err := dmChat.SendMessage(ctx, clientUsername, partnerUsername, msgContentJson, createdAt)
	if err != nil {
		return nil, err
	}

	go messageBrokerService.Send(fmt.Sprintf("user-%s-topic", partnerUsername), messageBrokerService.Message{
		Event: "new dm chat message",
		Data:  newMessage.PartnerNewMsgData,
	})

	return newMessage.ClientNewMsgData, nil
}

func AckMessageDelivered(ctx context.Context, clientUsername, partnerUsername, msgId string, deliveredAt time.Time) error {
	if err := dmChat.AckMessageDelivered(ctx, clientUsername, partnerUsername, msgId, deliveredAt); err != nil {
		return err
	}

	go messageBrokerService.Send(fmt.Sprintf("user-%s-topic", partnerUsername), messageBrokerService.Message{
		Event: "dm chat message delivered",
		Data: map[string]any{
			"partner_username": clientUsername,
			"msg_id":           msgId,
		},
	})
	return nil
}

func AckMessageRead(ctx context.Context, clientUsername, partnerUsername, msgId string, readAt time.Time) error {
	if err := dmChat.AckMessageRead(ctx, clientUsername, partnerUsername, msgId, readAt); err != nil {
		return err
	}
	go messageBrokerService.Send(fmt.Sprintf("user-%s-topic", partnerUsername), messageBrokerService.Message{
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
