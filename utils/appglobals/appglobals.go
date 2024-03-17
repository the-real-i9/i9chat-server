package appglobals

import (
	"errors"
)

var ErrInternalServerError = errors.New("internal server error: check logger")

var chatUpdatesObserver = make(map[string]chan<- map[string]any)

func SubscribeToChatUpdates(key string, mailbox chan<- map[string]any) {
	chatUpdatesObserver[key] = mailbox
}

func UnsubscribeFromChatUpdates(key string) {
	close(chatUpdatesObserver[key])
	delete(chatUpdatesObserver, key)
}
