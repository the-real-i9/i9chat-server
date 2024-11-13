package dmChatService

import (
	"fmt"
	"i9chat/appTypes"
	dmChat "i9chat/models/chatModel/dmChatModel"
	"i9chat/services/messageBrokerService"
	"i9chat/services/utils/appUtilServices"
	"time"
)

func GetChatHistory(dmChatId, offset int) ([]*dmChat.Message, error) {
	return dmChat.GetChatHistory(dmChatId, offset)
}

func SendMessage(dmChatId, senderId int, msgContent map[string]any, createdAt time.Time) (*dmChat.SenderData, error) {

	modMsgContent := appUtilServices.UploadMessageMedia(senderId, msgContent)

	newMessage, err := dmChat.SendMessage(dmChatId, senderId, modMsgContent, createdAt)
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

func NewDMChat(initiatorId, partnerId int, initMsgContent map[string]any, createdAt time.Time) (*dmChat.InitiatorData, error) {

	modInitMsgContent := appUtilServices.UploadMessageMedia(initiatorId, initMsgContent)

	dmChat, err := dmChat.New(initiatorId, partnerId, modInitMsgContent, createdAt)
	if err != nil {
		return nil, err
	}

	go messageBrokerService.PostMessage(fmt.Sprintf("user-%d", partnerId), messageBrokerService.Message{
		Event: "new dm chat",
		Data:  dmChat.PartnerData,
	})

	return dmChat.InitiatorData, nil
}

func UpdateMessageDeliveryStatus(dmChatId, msgId, senderId, receiverId int, status string, updatedAt time.Time) {
	if err := dmChat.UpdateMessageDeliveryStatus(dmChatId, msgId, receiverId, status, updatedAt); err == nil {

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

func BatchUpdateMessageDeliveryStatus(receiverId int, status string, ackDatas []*appTypes.DMChatMsgAckData) {
	if err := dmChat.BatchUpdateMessageDeliveryStatus(receiverId, status, ackDatas); err == nil {
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
