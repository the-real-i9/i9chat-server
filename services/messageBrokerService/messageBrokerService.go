package messageBrokerService

import (
	"context"
	"encoding/json"
	"fmt"
	"i9chat/appGlobals"
	"log"

	"github.com/segmentio/kafka-go"
)

type Message struct {
	Event string `json:"event" db:"event"`
	Data  any    `json:"data" db:"data"`
}

func Send(topic string, message Message) {
	msg, _ := json.Marshal(message)

	w := appGlobals.KafkaWriter

	err := w.WriteMessages(context.Background(), kafka.Message{
		Value: msg,
		Topic: fmt.Sprintf("i9chat-%s", topic),
	})

	if err != nil {
		log.Println("failed to write message:", err)
	}
}
