package appObservers

import (
	"fmt"
	"i9chat/utils/helpers"
	"sync"
)

type Observer interface {
	Subscribe(key string, mailbox chan<- map[string]any)
	Unsubscribe(key string)
	Send(key string, data any, event string)
}

// New Observers
// DMChatObserver - Events: ("new chat" | "new message"). dmChatId: (n)
// GroupChatObserver - Events: ("new chat" | "new message" | "new activity"). groupChatId: (n)
// DMChatActiveSessionObserver - Events: ("message update"). dmChatId: (n)
// GroupChatSessionObserver - Events: ("message update"). groupChatId: (n)

var (
	muDM           = sync.Mutex{}
	dmChatObserver = make(map[string]chan<- map[string]any)
)

type DMChatObserver struct{}

func (DMChatObserver) Subscribe(key string, mailbox chan<- map[string]any) {
	muDM.Lock()
	defer muDM.Unlock()

	dmChatObserver[key] = mailbox
}

func (DMChatObserver) Unsubscribe(key string) {
	muDM.Lock()
	defer muDM.Unlock()

	close(dmChatObserver[key])
	delete(dmChatObserver, key)
}

func (DMChatObserver) Send(key string, data any, event string) { // call in a new goroutine
	muDM.Lock()
	defer muDM.Unlock()

	if mailbox, found := dmChatObserver[key]; found {
		mailbox <- map[string]any{"event": event, "data": data}
	} else {
		var userId int
		fmt.Sscanf(key, "user-%d", &userId)

		go helpers.QueryRowField[bool](`
			INSERT INTO dm_chat_event_pending_receipt (user_id, event, data) 
			VALUES ($1, $2, $3) 
			RETURNING true
		`, userId, event, data)

	}
}

func (DMChatObserver) SendPresenceUpdate(key string, data any, event string) {
	muDM.Lock()
	defer muDM.Unlock()

	if mailbox, found := dmChatObserver[key]; found {
		mailbox <- map[string]any{"event": event, "data": data}
	}
}

var (
	muGroup           = sync.Mutex{}
	groupChatObserver = make(map[string]chan<- map[string]any)
)

type GroupChatObserver struct{}

func (GroupChatObserver) Subscribe(key string, mailbox chan<- map[string]any) {
	muGroup.Lock()
	defer muGroup.Unlock()

	groupChatObserver[key] = mailbox
}

func (GroupChatObserver) Unsubscribe(key string) {
	muGroup.Lock()
	defer muGroup.Unlock()

	close(groupChatObserver[key])
	delete(groupChatObserver, key)
}

func (GroupChatObserver) Send(key string, data any, event string) { // call in a new goroutine
	muGroup.Lock()
	defer muGroup.Unlock()

	if mailbox, found := groupChatObserver[key]; found {
		mailbox <- map[string]any{"event": event, "data": data}
	} else {
		var userId int
		fmt.Sscanf(key, "user-%d", &userId)

		go helpers.QueryRowField[bool](`
			INSERT INTO group_chat_event_pending_receipt (user_id, event, data) 
			VALUES ($1, $2, $3) 
			RETURNING true
		`, userId, event, data)
	}
}

var (
	muDMSess              = sync.Mutex{}
	dmChatSessionObserver = make(map[string]chan<- map[string]any)
)

type DMChatSessionObserver struct{}

func (DMChatSessionObserver) Subscribe(key string, mailbox chan<- map[string]any) {
	muDMSess.Lock()
	defer muDMSess.Unlock()

	dmChatSessionObserver[key] = mailbox
}

func (DMChatSessionObserver) Unsubscribe(key string) {
	muDMSess.Lock()
	defer muDMSess.Unlock()

	close(dmChatSessionObserver[key])
	delete(dmChatSessionObserver, key)
}

func (DMChatSessionObserver) Send(key string, data any, event string) { // call in a new goroutine
	muDMSess.Lock()
	defer muDMSess.Unlock()

	if mailbox, found := dmChatSessionObserver[key]; found {
		mailbox <- map[string]any{"event": event, "data": data}
	} else {
		var (
			userId   int
			dmChatId int
		)

		fmt.Sscanf(key, "user-%d--dmchat-%d", &userId, &dmChatId)

		go helpers.QueryRowField[bool](`
			INSERT INTO dm_chat_message_event_pending_receipt (user_id, dm_chat_id, event, data) 
			VALUES ($1, $2, $3, $4) 
			RETURNING true
		`, userId, dmChatId, event, data)
	}
}

var (
	muGroupSess              = sync.Mutex{}
	groupChatSessionObserver = make(map[string]chan<- map[string]any)
)

type GroupChatSessionObserver struct{}

func (GroupChatSessionObserver) Subscribe(key string, mailbox chan<- map[string]any) {
	muGroupSess.Lock()
	defer muGroupSess.Unlock()

	groupChatSessionObserver[key] = mailbox
}

func (GroupChatSessionObserver) Unsubscribe(key string) {
	muGroupSess.Lock()
	defer muGroupSess.Unlock()

	close(groupChatSessionObserver[key])
	delete(groupChatSessionObserver, key)
}

func (GroupChatSessionObserver) Send(key string, data any, event string) { // call in a new goroutine
	muGroupSess.Lock()
	defer muGroupSess.Unlock()

	if mailbox, found := groupChatSessionObserver[key]; found {
		mailbox <- map[string]any{"event": event, "data": data}
	} else {
		var (
			userId      int
			groupChatId int
		)

		fmt.Sscanf(key, "user-%d--groupchat-%d", &userId, &groupChatId)

		go helpers.QueryRowField[bool](`
			INSERT INTO group_chat_message_event_pending_receipt (user_id, group_chat_id, event, data) 
			VALUES ($1, $2, $3, $4) 
			RETURNING true
		`, userId, groupChatId, event, data)
	}
}
