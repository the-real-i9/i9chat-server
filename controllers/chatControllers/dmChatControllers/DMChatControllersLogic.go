package dmChatControllers

import (
	"fmt"
	dmChat "i9chat/models/chatModel/dmChatModel"
	"i9chat/services/messageBrokerService"
	"time"
)

func sendMessage(dmChatId, senderId int, msgContent map[string]any, createdAt time.Time) (*dmChat.SenderData, error) {

	newMessage, err := dmChat.SendMessage(dmChatId, senderId, msgContent, createdAt)
	if err != nil {
		return nil, err
	}

	go messageBrokerService.PostMessage(fmt.Sprintf("user-%d", newMessage.ReceiverId), messageBrokerService.Message{
		Event: "new dm chat message",
		Data:  newMessage.ReceiverData,
	})

	return newMessage.SenderData, nil
}

func react() {

}
