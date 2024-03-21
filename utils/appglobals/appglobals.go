package appglobals

import (
	"context"
	"errors"
	"os"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

var ErrInternalServerError = errors.New("internal server error: check logger")

var GCSClient *storage.Client

func InitGCSClient() error {
	stClient, err := storage.NewClient(context.Background(), option.WithCredentialsFile(os.Getenv("GCS_CRED_FILE")))
	if err != nil {
		return err
	}

	GCSClient = stClient

	return nil
}

type Observer interface {
	Subscribe(key string, mailbox chan<- map[string]any)
	Unsubscribe(key string)
	Send(key string, data map[string]any, event string)
}

var chatObserver = make(map[string]chan<- map[string]any)

type ChatObserver struct{}

func (ChatObserver) Subscribe(key string, mailbox chan<- map[string]any) {
	chatObserver[key] = mailbox
}

func (ChatObserver) Unsubscribe(key string) {
	close(chatObserver[key])
	delete(chatObserver, key)
}

func (ChatObserver) Send(key string, data map[string]any, event string) { // call in a new goroutine
	if mailbox, found := chatObserver[key]; found {
		mailbox <- map[string]any{"event": event, "data": data}
	}
}

var dMChatMessageObserver = make(map[string]chan<- map[string]any)

type DMChatMessageObserver struct{}

func (DMChatMessageObserver) Subscribe(key string, mailbox chan<- map[string]any) {
	dMChatMessageObserver[key] = mailbox
}

func (DMChatMessageObserver) Unsubscribe(key string) {
	close(dMChatMessageObserver[key])
	delete(dMChatMessageObserver, key)
}

func (DMChatMessageObserver) Send(key string, data map[string]any, event string) { // call in a new goroutine
	if mailbox, found := dMChatMessageObserver[key]; found {
		mailbox <- map[string]any{"event": event, "data": data}
	}
}

var groupChatMessageObserver = make(map[string]chan<- map[string]any)

type GroupChatMessageObserver struct{}

func (GroupChatMessageObserver) Subscribe(key string, mailbox chan<- map[string]any) {
	groupChatMessageObserver[key] = mailbox
}

func (GroupChatMessageObserver) Unsubscribe(key string) {
	close(groupChatMessageObserver[key])
	delete(groupChatMessageObserver, key)
}

func (GroupChatMessageObserver) Send(key string, data map[string]any, event string) { // call in a new goroutine
	if mailbox, found := groupChatMessageObserver[key]; found {
		mailbox <- map[string]any{"event": event, "data": data}
	}
}

var groupChatActivityObserver = make(map[string]chan<- map[string]any)

type GroupChatActivityObserver struct{}

func (GroupChatActivityObserver) Subscribe(key string, mailbox chan<- map[string]any) {
	groupChatActivityObserver[key] = mailbox
}

func (GroupChatActivityObserver) Unsubscribe(key string) {
	close(groupChatActivityObserver[key])
	delete(groupChatActivityObserver, key)
}

func (GroupChatActivityObserver) Log(key string, data map[string]any) { // call in a new goroutine
	if mailbox, found := groupChatActivityObserver[key]; found {
		mailbox <- data
	}
}
