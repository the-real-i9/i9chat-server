package eventStreamService

import (
	"context"
	"encoding/json"
	"i9chat/src/appTypes"
	"i9chat/src/models/db"
	"log"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var usersEventPipes = &sync.Map{}

func Send(receiverUser string, message appTypes.ServerWSMsg) {
	if userPipe, ok := usersEventPipes.Load(receiverUser); ok {
		pipe := userPipe.(*websocket.Conn)

		if err := pipe.WriteJSON(message); err != nil {
			log.Println("error: eventStreamService.go: Send: pipe.WriteJSON", err)
		}

		return
	}

	strMsg, err := json.Marshal(message)
	if err != nil {
		log.Println("error: eventStreamService.go: Send: json.Marshal:", err)
		return
	}

	storeUndeliveredEventMsg(receiverUser, strMsg)
}

func Subscribe(clientUsername string, pipe *websocket.Conn) {
	undEventMsgs := getUndeliveredEventMsgs(clientUsername)

	for _, eventMsg := range undEventMsgs {
		eventMsg := eventMsg.(map[string]any)

		var msg appTypes.ServerWSMsg

		if err := json.Unmarshal(eventMsg["msg"].([]byte), &msg); err != nil {
			log.Println("error: eventStreamService.go: Subscribe: json.Unmarshal", err)
			return
		}

		if err := pipe.WriteJSON(msg); err != nil {
			log.Println("error: eventStreamService.go: Subscribe: pipe.WriteJSON", err)
			return
		}

		go deleteDeliveredEventMsg(clientUsername, eventMsg["id"].(string))
	}

	// add user pipe to user event pipes
	usersEventPipes.Store(clientUsername, pipe)
}

func Unsubscribe(clientUsername string) {
	go usersEventPipes.Delete(clientUsername)
}

func storeUndeliveredEventMsg(username string, msg []byte) {
	_, err := db.Query(
		context.Background(),
		`
		MATCH (clientUser:User{ username: $username })

		CREATE (clientUser)-[:HAS_UNDELIVERED_EVENT_MSG { at: $at }]->(:UndEventMessage{ id: randomUUID(), msg: $msg })
		`,
		map[string]any{
			"username": username,
			"msg":      msg,
			"at":       time.Now().UTC(),
		},
	)
	if err != nil {
		log.Println("eventStreamService.go: storeUndeliveredEventMsg:", err)
	}
}

func deleteDeliveredEventMsg(username, msgId string) {
	_, err := db.Query(
		context.Background(),
		`
		MATCH (clientUser:User{ username: $username })-[:HAS_UNDELIVERED_EVENT_MSG]->(dEventMsg:UndEventMessage{ id: $msgId })

		DETACH DELETE dEventMsg
		`,
		map[string]any{
			"username": username,
			"msgId":    msgId,
			"at":       time.Now().UTC(),
		},
	)
	if err != nil {
		log.Println("eventStreamService.go: deleteDeliveredEventMsg:", err)
	}
}

func getUndeliveredEventMsgs(username string) []any {
	res, err := db.Query(
		context.Background(),
		`
		MATCH (clientUser:User{ username: $username })-[rel:HAS_UNDELIVERED_EVENT_MSG]->(undEventMsg:UndEventMessage)

		ORDER BY rel.at ASC

		RETURN collect(undEventMsg { .* }) AS und_event_msgs
		`,
		map[string]any{
			"username": username,
		},
	)
	if err != nil {
		log.Println("eventStreamService.go: getUndeliveredEventMsgs:", err)
		return nil
	}

	if len(res.Records) == 0 {
		return nil
	}

	undEventMsgs, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "und_event_msgs")

	return undEventMsgs
}
