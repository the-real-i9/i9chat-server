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

var newMessageObserver = make(map[string]chan<- map[string]any)

func SubscribeToNewMessageUpdate(key string, mailbox chan<- map[string]any) {
	newMessageObserver[key] = mailbox
}

func UnsubscribeFromNewMessageUpdate(key string) {
	close(newMessageObserver[key])
	delete(newMessageObserver, key)
}

func SendNewChatMessageUpdate(key string, data map[string]any) { // call in a new goroutine
	if mailbox, found := newMessageObserver[key]; found {
		mailbox <- data
	}
}
