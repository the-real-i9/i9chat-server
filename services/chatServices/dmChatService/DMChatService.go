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

func GetChatHistory(ctx context.Context, dmChatId string, offset int) ([]*dmChat.Message, error) {
	return dmChat.GetChatHistory(ctx, dmChatId, offset)
}

func SendMessage(ctx context.Context, clientUserId, partnerUserId int, msgContent *appTypes.MsgContent, createdAt time.Time) (*dmChat.ClientNewMsgData, error) {

	err := appServices.UploadMessageMedia(ctx, clientUserId, msgContent)
	if err != nil {
		return nil, err
	}

	newMessage, err := dmChat.SendMessage(ctx, clientUserId, partnerUserId, *msgContent, createdAt)
	if err != nil {
		return nil, err
	}

	go messageBrokerService.Send(fmt.Sprintf("user-%d-topic", partnerUserId), messageBrokerService.Message{
		Event: "new dm chat message",
		Data:  newMessage.PartnerNewMsgData,
	})

	respData := newMessage.ClientNewMsgData

	return respData, nil
}

func UpdateMessageDeliveryStatus(ctx context.Context, clientUserId, partnerUserId, msgId int, status string, updatedAt time.Time) {
	if err := dmChat.UpdateMessageDeliveryStatus(ctx, clientUserId, partnerUserId, msgId, status, updatedAt); err == nil {

		go messageBrokerService.Send(fmt.Sprintf("user-%d-topic", partnerUserId), messageBrokerService.Message{
			Event: "dm chat message delivery status changed",
			Data: map[string]any{
				"partnerUserId": clientUserId,
				"msgId":         msgId,
				"status":        status,
			},
		})
	}
}

func BatchUpdateMessageDeliveryStatus(ctx context.Context, clientUserId int, status string, ackDatas []*appTypes.DMChatMsgAckData) {
	if err := dmChat.BatchUpdateMessageDeliveryStatus(ctx, clientUserId, status, ackDatas); err == nil {

		for _, data := range ackDatas {
			data := data

			messageBrokerService.Send(fmt.Sprintf("user-%d-topic", data.PartnerUserId), messageBrokerService.Message{
				Event: "dm chat message delivery status changed",
				Data: map[string]any{
					"partnerUserId": clientUserId,
					"msgId":         data.MsgId,
					"status":        status,
				},
			})
		}
	}
}

func React() {

}
