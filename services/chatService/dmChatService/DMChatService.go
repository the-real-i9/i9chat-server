package dmChatService

import (
	"fmt"
	dmChat "i9chat/models/chatModel/dmChatModel"
	"i9chat/services/appObservers"
	"time"
)

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

func React() {

}
