package messageBrokerService

import (
	"fmt"
	"i9chat/helpers"
	"sync"
)

// I personally designed this message broker, but I don't know which known pattern it follows,
// so I call it the "post office" pattern

// The broker will have the following methods: AddMailbox, RemoveMailbox, and Send
// AddMailbox():
// The user calls whis method when he wants to start receiveing messages/events from the server, practically by "coming online".
// On call: It accepts a user's mailbox in which it wants to receive messages, and adds it to a map of user mailboxes in the post office.
// The method will use the user's PostOffice(PO) id to check for "must-deliver" messages pending delivery
// as queued in the database, and stream it to the user's mailbox

// RemoveMailbox():
// This user calls this method when he wants to stop receiving messages/events from the server, practically by "going offline".
// On call: It accepts a user's PostOffice(PO) id and uses it to remove their mailbox from the map of user mailboxes.

// PostMessage():
// This method is called when a user wants to post a message/event to another user
// On call: It accepts a user's PostOffice(PO) id whose mailbox the sender wants to post a message
// If the user's mailbox is missing in the map of user mailboxes,
// a "must-deliver" flag, set to true, causes this message to be persisted in the database
// alongside the receiver's user PostOffice(PO) id.

var (
	mu         = sync.Mutex{}
	postOffice = make(map[string]chan<- any)
)

type Message struct {
	Event string `json:"event" db:"event"`
	Data  any    `json:"data" db:"data"`
}

func AddMailbox(userPOId string, mailbox chan<- any) {
	mu.Lock()
	defer mu.Unlock()

	if _, found := postOffice[userPOId]; found {
		return
	}

	postOffice[userPOId] = mailbox

	streamMessagesPendingDelivery(userPOId, mailbox)

}

func streamMessagesPendingDelivery(userPOId string, mailbox chan<- any) {
	var userId int
	fmt.Sscanf(userPOId, "user-%d", &userId)

	messages, err := helpers.QueryRowsField[Message](`SELECT * FROM fetch_user_broker_messages_pending_delivery($1)`, userId)
	if err != nil {
		panic(err)
	}

	for _, msg := range messages {
		msg := *msg
		mailbox <- msg
	}
}

func RemoveMailbox(userPOId string) {
	mu.Lock()
	defer mu.Unlock()

	if _, found := postOffice[userPOId]; !found {
		return
	}

	close(postOffice[userPOId])
	delete(postOffice, userPOId)
}

func PostMessage(userPOId string, msg Message) {
	mu.Lock()
	defer mu.Unlock()

	if mailbox, found := postOffice[userPOId]; found {
		mailbox <- msg
	} else {
		var userId int
		fmt.Sscanf(userPOId, "user-%d", &userId)

		go helpers.QueryRowField[bool](`
			INSERT INTO user_broker_message_pending_delivery (user_id, message) 
			VALUES ($1, $2) 
			RETURNING true
		`, userId, msg)

	}
}
