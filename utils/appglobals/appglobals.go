package appglobals

import (
	"errors"
	"fmt"
	"strings"
)

var ErrInternalServerError = errors.New("internal server error: check logger")

var newChatObserver = make(map[string]chan<- map[string]any)

func SubscribeToNewChatUpdate(key string, mailbox chan<- map[string]any) {
	newChatObserver[key] = mailbox
}

func UnsubscribeFromNewChatUpdate(key string) {
	close(newChatObserver[key])
	delete(newChatObserver, key)
}

func SendNewChatUpdate(key string, data map[string]any) { // call in a new goroutine
	if mailbox, found := newChatObserver[key]; found {
		mailbox <- data
	}
}

var newDMChatMessageObserver = make(map[string]chan<- map[string]any)

func SubscribeToNewDMChatMessageUpdate(key string, mailbox chan<- map[string]any) {
	newDMChatMessageObserver[key] = mailbox
}

func UnsubscribeFromNewDMChatMessageUpdate(key string) {
	close(newDMChatMessageObserver[key])
	delete(newDMChatMessageObserver, key)
}

func SendNewDMChatMessageUpdate(chatId int, skipUserId int, data map[string]any) { // call in a new goroutine
	for key, mailbox := range newDMChatMessageObserver {
		if !strings.Contains(key, fmt.Sprintf("dmchat-%d", chatId)) ||
			strings.Contains(key, fmt.Sprintf("user-%d", skipUserId)) {
			continue
		}

		go func(mailbox chan<- map[string]any) {
			mailbox <- data
		}(mailbox)
	}
}

var newGroupChatMessageObserver = make(map[string]chan<- map[string]any)

func SubscribeToNewGroupChatMessageUpdate(key string, mailbox chan<- map[string]any) {
	newGroupChatMessageObserver[key] = mailbox
}

func UnsubscribeFromNewGroupChatMessageUpdate(key string) {
	close(newGroupChatMessageObserver[key])
	delete(newGroupChatMessageObserver, key)
}

func SendNewGroupChatMessageUpdate(chatId int, skipUserId int, data map[string]any) { // call in a new goroutine
	for key, mailbox := range newGroupChatMessageObserver {
		if !strings.Contains(key, fmt.Sprintf("groupchat-%d", chatId)) ||
			strings.Contains(key, fmt.Sprintf("user-%d", skipUserId)) {
			continue
		}

		go func(mailbox chan<- map[string]any) {
			mailbox <- data
		}(mailbox)
	}
}
