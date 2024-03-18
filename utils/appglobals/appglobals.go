package appglobals

import (
	"errors"
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

func SendNewDMChatMessageUpdate(key string, data map[string]any) { // call in a new goroutine
	if mailbox, found := newDMChatMessageObserver[key]; found {
		mailbox <- data
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

func SendNewGroupChatMessageUpdate(key string, data map[string]any) { // call in a new goroutine
	if mailbox, found := newGroupChatMessageObserver[key]; found {
		mailbox <- data
	}
}
