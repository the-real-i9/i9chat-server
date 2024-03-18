package appglobals

import (
	"errors"
)

var ErrInternalServerError = errors.New("internal server error: check logger")

type Observer interface {
	Subscribe(key string, mailbox chan<- map[string]any)
	Unubscribe(key string)
	Send(key string, data map[string]any)
}

var newChatObserver = make(map[string]chan<- map[string]any)

type NewChatObserver struct{}

func (NewChatObserver) Subscribe(key string, mailbox chan<- map[string]any) {
	newChatObserver[key] = mailbox
}

func (NewChatObserver) Unsubscribe(key string) {
	close(newChatObserver[key])
	delete(newChatObserver, key)
}

func (NewChatObserver) Send(key string, data map[string]any) { // call in a new goroutine
	if mailbox, found := newChatObserver[key]; found {
		mailbox <- data
	}
}

var chatUpdateObserver = make(map[string]chan<- map[string]any)

type ChatUpdateObserver struct{}

func (ChatUpdateObserver) Subscribe(key string, mailbox chan<- map[string]any) {
	chatUpdateObserver[key] = mailbox
}

func (ChatUpdateObserver) Unsubscribe(key string) {
	close(chatUpdateObserver[key])
	delete(chatUpdateObserver, key)
}

func (ChatUpdateObserver) Send(key string, data map[string]any) { // call in a new goroutine
	if mailbox, found := chatUpdateObserver[key]; found {
		mailbox <- data
	}
}

var newDMChatMessageObserver = make(map[string]chan<- map[string]any)

type NewDMChatMessageObserver struct{}

func (NewDMChatMessageObserver) Subscribe(key string, mailbox chan<- map[string]any) {
	newDMChatMessageObserver[key] = mailbox
}

func (NewDMChatMessageObserver) Unsubscribe(key string) {
	close(newDMChatMessageObserver[key])
	delete(newDMChatMessageObserver, key)
}

func (NewDMChatMessageObserver) Send(key string, data map[string]any) { // call in a new goroutine
	if mailbox, found := newDMChatMessageObserver[key]; found {
		mailbox <- data
	}
}

var dmChatMessageUpdateObserver = make(map[string]chan<- map[string]any)

type DmChatMessageUpdateObserver struct{}

func (DmChatMessageUpdateObserver) Subscribe(key string, mailbox chan<- map[string]any) {
	dmChatMessageUpdateObserver[key] = mailbox
}

func (DmChatMessageUpdateObserver) Unubscribe(key string) {
	close(dmChatMessageUpdateObserver[key])
	delete(dmChatMessageUpdateObserver, key)
}

func (DmChatMessageUpdateObserver) Send(key string, data map[string]any) { // call in a new goroutine
	if mailbox, found := dmChatMessageUpdateObserver[key]; found {
		mailbox <- data
	}
}

var newGroupChatMessageObserver = make(map[string]chan<- map[string]any)

type NewGroupChatMessageObserver struct{}

func (NewGroupChatMessageObserver) Subscribe(key string, mailbox chan<- map[string]any) {
	newGroupChatMessageObserver[key] = mailbox
}

func (NewGroupChatMessageObserver) Unubscribe(key string) {
	close(newGroupChatMessageObserver[key])
	delete(newGroupChatMessageObserver, key)
}

func (NewGroupChatMessageObserver) Send(key string, data map[string]any) { // call in a new goroutine
	if mailbox, found := newGroupChatMessageObserver[key]; found {
		mailbox <- data
	}
}

var groupChatMessageUpdateObserver = make(map[string]chan<- map[string]any)

type GroupChatMessageUpdateObserver struct{}

func (GroupChatMessageUpdateObserver) Subscribe(key string, mailbox chan<- map[string]any) {
	groupChatMessageUpdateObserver[key] = mailbox
}

func (GroupChatMessageUpdateObserver) Unsubscribe(key string) {
	close(groupChatMessageUpdateObserver[key])
	delete(groupChatMessageUpdateObserver, key)
}

func (GroupChatMessageUpdateObserver) Send(key string, data map[string]any) { // call in a new goroutine
	if mailbox, found := groupChatMessageUpdateObserver[key]; found {
		mailbox <- data
	}
}
