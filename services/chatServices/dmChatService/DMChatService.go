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

func NewDMChat(ctx context.Context, clientUserId, partnerUserId int, initMsgContent *appTypes.MsgContent, createdAt time.Time) (*dmChat.ClientNewDMChatData, error) {

	err := appServices.UploadMessageMedia(ctx, clientUserId, initMsgContent)
	if err != nil {
		return nil, err
	}

	dmChat, err := dmChat.New(ctx, clientUserId, partnerUserId, *initMsgContent, createdAt)
	if err != nil {
		return nil, err
	}

	go messageBrokerService.Send(fmt.Sprintf("user-%d-topic", partnerUserId), messageBrokerService.Message{
		Event: "new dm chat",
		Data:  dmChat.PartnerNewDMChatData,
	})

	return dmChat.ClientNewDMChatData, nil
}

func GetChatHistory(ctx context.Context, dmChatId string, offset int) ([]*dmChat.Message, error) {
	return dmChat.GetChatHistory(ctx, dmChatId, offset)
}

func SendMessage(ctx context.Context, clientDMChatId string, clientUserId int, msgContent *appTypes.MsgContent, createdAt time.Time) (*dmChat.ClientNewMsgData, error) {

	err := appServices.UploadMessageMedia(ctx, clientUserId, msgContent)
	if err != nil {
		return nil, err
	}

	newMessage, err := dmChat.SendMessage(ctx, clientDMChatId, clientUserId, *msgContent, createdAt)
	if err != nil {
		return nil, err
	}

	go messageBrokerService.Send(fmt.Sprintf("user-%d-topic", newMessage.PartnerUserId), messageBrokerService.Message{
		Event: "new dm chat message",
		Data:  newMessage.PartnerNewMsgData,
	})

	respData := newMessage.ClientNewMsgData

	return respData, nil
}

func UpdateMessageDeliveryStatus(ctx context.Context, clientDMChatId string, msgId, clientUserId int, status string, updatedAt time.Time) {
	if res, err := dmChat.UpdateMessageDeliveryStatus(ctx, clientDMChatId, msgId, clientUserId, status, updatedAt); err == nil {

		go messageBrokerService.Send(fmt.Sprintf("user-%d-topic", res.PartnerUserId), messageBrokerService.Message{
			Event: "dm chat message delivery status changed",
			Data: map[string]any{
				"dmChatId": res.PartnerDMChatId,
				"msgId":    res.MsgId,
				"status":   status,
			},
		})
	}
}

func BatchUpdateMessageDeliveryStatus(ctx context.Context, clientUserId int, status string, ackDatas []*appTypes.DMChatMsgAckData) {
	if res, err := dmChat.BatchUpdateMessageDeliveryStatus(ctx, clientUserId, status, ackDatas); err == nil {

		for _, data := range res {
			data := data

			messageBrokerService.Send(fmt.Sprintf("user-%d-topic", data.PartnerUserId), messageBrokerService.Message{
				Event: "dm chat message delivery status changed",
				Data: map[string]any{
					"dmChatId": data.PartnerDMChatId,
					"msgId":    data.MsgId,
					"status":   status,
				},
			})
		}
	}
}

func React() {

}
