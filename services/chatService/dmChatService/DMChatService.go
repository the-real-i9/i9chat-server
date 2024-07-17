package dmChatService

import (
	"fmt"
	dmChat "i9chat/models/chatModel/dmChatModel"
	"i9chat/services/appObservers"
	"i9chat/utils/appTypes"
	"time"
)

func NewDMChat(initiatorId, partnerId int, initMsgContent map[string]any, createdAt time.Time) (*dmChat.InitiatorData, error) {
	dmChat, err := dmChat.New(initiatorId, partnerId, initMsgContent, createdAt)
	if err != nil {
		return nil, err
	}

	go appObservers.DMChatObserver{}.Send(fmt.Sprintf("user-%d", partnerId), dmChat.PartnerData, "new chat")

	return dmChat.InitiatorData, nil
}

func BatchUpdateMessageDeliveryStatus(receiverId int, status string, ackDatas []*appTypes.DMChatMsgAckData) {
	if err := dmChat.BatchUpdateMessageDeliveryStatus(receiverId, status, ackDatas); err == nil {
		for _, data := range ackDatas {
			data := data
			go func() {
				appObservers.DMChatSessionObserver{}.Send(
					fmt.Sprintf("user-%d--dmchat-%d", data.SenderId, data.DMChatId),
					map[string]any{"msgId": data.MsgId, "status": status},
					"delivery status update",
				)
			}()
		}
	}
}

func SendMessage(dmChatId, senderId int, msgContent map[string]any, createdAt time.Time) (*dmChat.SenderData, error) {

	newMessage, err := dmChat.SendMessage(dmChatId, senderId, msgContent, createdAt)
	if err != nil {
		return nil, err
	}

	go appObservers.DMChatObserver{}.Send(
		fmt.Sprintf("user-%d", newMessage.ReceiverId), newMessage.ReceiverData, "new message",
	)

	return newMessage.SenderData, nil
}

func UpdateMessageDeliveryStatus(dmChatId, msgId, senderId, receiverId int, status string, updatedAt time.Time) {
	if err := dmChat.UpdateMessageDeliveryStatus(dmChatId, msgId, receiverId, status, updatedAt); err == nil {

		go appObservers.DMChatSessionObserver{}.Send(
			fmt.Sprintf("user-%d--dmchat-%d", senderId, dmChatId),
			map[string]any{"msgId": msgId, "status": status},
			"delivery status update",
		)
	}

}

func React() {

}
