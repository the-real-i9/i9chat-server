package messageBrokerService

import (
	"context"
	"encoding/json"
	"fmt"
	"i9chat/appGlobals"
	"log"
	"os"

	"github.com/segmentio/kafka-go"
)

type Message struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
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

func ConsumeTopic(topic string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{os.Getenv("KAFKA_BROKER_ADDRESS")},
		Topic:   "i9chat-" + topic,
		GroupID: "i9chat-topics",
		// CommitInterval: time.Second,
	})
}
