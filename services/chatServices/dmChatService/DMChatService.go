package dmChatService

import (
	"context"
	"fmt"
	"i9chat/appTypes"
	dmChat "i9chat/models/chatModel/dmChatModel"
	"i9chat/services/appServices"
	"i9chat/services/messageBrokerService"
	"time"
)

func NewDMChat(ctx context.Context, initiatorId, partnerId int, initMsgContent map[string]any, createdAt time.Time) (*dmChat.InitiatorData, error) {

	modInitMsgContent, err := appServices.UploadMessageMedia(ctx, initiatorId, initMsgContent)
	if err != nil {
		return nil, err
	}

	dmChat, err := dmChat.New(ctx, initiatorId, partnerId, modInitMsgContent, createdAt)
	if err != nil {
		return nil, err
	}

	go messageBrokerService.PostMessage(fmt.Sprintf("user-%d", partnerId), messageBrokerService.Message{
		Event: "new dm chat",
		Data:  dmChat.PartnerData,
	})

	return dmChat.InitiatorData, nil
}

func GetChatHistory(ctx context.Context, dmChatId, offset int) ([]*dmChat.Message, error) {
	return dmChat.GetChatHistory(ctx, dmChatId, offset)
}

func SendMessage(ctx context.Context, dmChatId, senderId int, msgContent map[string]any, createdAt time.Time) (*dmChat.SenderData, error) {

	modMsgContent, err := appServices.UploadMessageMedia(ctx, senderId, msgContent)
	if err != nil {
		return nil, err
	}

	newMessage, err := dmChat.SendMessage(ctx, dmChatId, senderId, modMsgContent, createdAt)
	if err != nil {
		return nil, err
	}

	go messageBrokerService.PostMessage(fmt.Sprintf("user-%d", newMessage.ReceiverId), messageBrokerService.Message{
		Event: "new dm chat message",
		Data:  newMessage.ReceiverData,
	})

	respData := newMessage.SenderData

	return respData, nil
}

func UpdateMessageDeliveryStatus(ctx context.Context, dmChatId, msgId, senderId, receiverId int, status string, updatedAt time.Time) {
	if err := dmChat.UpdateMessageDeliveryStatus(ctx, dmChatId, msgId, receiverId, status, updatedAt); err == nil {

		go messageBrokerService.PostMessage(fmt.Sprintf("user-%d", senderId), messageBrokerService.Message{
			Event: "dm chat message delivery status changed",
			Data: map[string]any{
				"dmChatId": dmChatId,
				"msgId":    msgId,
				"status":   status,
			},
		})
	}
}

func BatchUpdateMessageDeliveryStatus(ctx context.Context, receiverId int, status string, ackDatas []*appTypes.DMChatMsgAckData) {
	if err := dmChat.BatchUpdateMessageDeliveryStatus(ctx, receiverId, status, ackDatas); err == nil {
		for _, data := range ackDatas {
			data := data

			go messageBrokerService.PostMessage(fmt.Sprintf("user-%d", data.SenderId), messageBrokerService.Message{
				Event: "dm chat message delivery status changed",
				Data: map[string]any{
					"dmChatId": data.DMChatId,
					"msgId":    data.MsgId,
					"status":   status,
				},
			})
		}
	}
}

func React() {

}
