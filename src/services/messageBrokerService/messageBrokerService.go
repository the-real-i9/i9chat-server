package messageBrokerService

import (
	"context"
	"encoding/json"
	"fmt"
	"i9chat/src/appGlobals"
	"log"
	"net"
	"os"
	"strconv"

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

	createTopic(topic)

	return kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{os.Getenv("KAFKA_BROKER_ADDRESS")},
		Topic:   "i9chat-" + topic,
		GroupID: "i9chat-topics",
		// CommitInterval: time.Second,
	})
}

func createTopic(topic string) {

	topic = "i9chat-" + topic

	conn, err := kafka.Dial("tcp", os.Getenv("KAFKA_BROKER_ADDRESS"))
	if err != nil {
		log.Printf("messageBrokerService.go: CreateTopic(%s): kafka.Dial: %s", topic, err)
		return
	}

	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		log.Printf("messageBrokerService.go: CreateTopic(%s): conn.Controller: %s", topic, err)
		return
	}
	var controllerConn *kafka.Conn
	controllerConn, err = kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		log.Printf("messageBrokerService.go: CreateTopic(%s): kafka.Dial(2): %s", topic, err)
		return
	}
	defer controllerConn.Close()

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	}

	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		log.Printf("messageBrokerService.go: CreateTopic(%s): CreateTopics: %s", topic, err)
		return
	}
}
