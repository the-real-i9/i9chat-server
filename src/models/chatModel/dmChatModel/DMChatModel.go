package dmChat

import (
	"context"
	"i9chat/src/helpers"
	"i9chat/src/models/db"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type NewMessage struct {
	ClientData  map[string]any `json:"client_resp"`
	PartnerData map[string]any `json:"partner_resp"`
}

func SendMessage(ctx context.Context, clientUsername, partnerUsername, msgContent string, createdAt time.Time) (NewMessage, error) {
	var newMsg NewMessage

	res, err := db.Query(
		ctx,
		`
		MATCH (clientUser:User{ username: $client_username }), (partnerUser:User{ username: $partner_username })
		MERGE (clientUser)-[:HAS_CHAT]->(clientChat:DMChat{ owner_username: $client_username, partner_username: $partner_username })-[:WITH_USER]->(partnerUser)
		MERGE (partnerUser)-[:HAS_CHAT]->(partnerChat:DMChat{ owner_username: $partner_username, partner_username: $client_username })-[:WITH_USER]->(clientUser)
		SET clientChat.last_activity_type = "message", 
			partnerChat.last_activity_type = "message",
			clientChat.updated_at = $created_at, 
			partnerChat.updated_at = $created_at

		WITH clientUser, clientChat, partnerUser, partnerChat
		CREATE (message:DMMessage{ id: randomUUID(), content: $message_content, delivery_status: "sent", created_at: $created_at }),
			(clientUser)-[:SENDS_MESSAGE]->(message)-[:IN_DM_CHAT]->(clientChat),
			(partnerUser)-[:RECEIVES_MESSAGE]->(message)-[:IN_DM_CHAT]->(partnerChat)
		SET clientChat.last_message_id = message.id,
			partnerChat.last_message_id = message.id
			
		WITH message, toString(message.created_at) AS created_at, clientUser { .username, .profile_pic_url, .connection_status } AS sender
		RETURN { new_msg_id: message.id } AS client_resp,
			message { .*, created_at, sender } AS partner_resp
		`,
		map[string]any{
			"client_username":  clientUsername,
			"partner_username": partnerUsername,
			"message_content":  msgContent,
			"created_at":       createdAt,
		},
	)
	if err != nil {
		log.Println("DMChatModel.go: SendMessage:", err)
		return newMsg, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return newMsg, nil
	}

	helpers.ToStruct(res.Records[0].AsMap(), &newMsg)

	return newMsg, nil
}

func ReactToMessage(ctx context.Context, clientUsername, msgId string, reaction rune) error {

	return nil
}

func ChatHistory(ctx context.Context, clientUsername, partnerUsername string, limit int, offset time.Time) ([]any, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (clientChat:DMChat{ owner_username: $client_username, partner_username: $partner_username })
		
		OPTIONAL MATCH (clientChat)<-[:IN_DM_CHAT]-(message:DMMessage WHERE message.created_at < $offset)
		OPTIONAL MATCH (message)<-[:SENDS_MESSAGE]-(senderUser)
		OPTIONAL MATCH (message)<-[rxn:REACTS_TO_MESSAGE]-(reactorUser)
			
		WITH message, toString(message.created_at) AS created_at, senderUser { .username, .profile_pic_url } AS sender, collect({ user: reactorUser { .username, .profile_pic_url }, reaction: rxn.reaction }) AS reactions
		ORDER BY message.created_at
		LIMIT $limit
		RETURN collect(message { .*, created_at, sender, reactions }) AS chat_history
		`,
		map[string]any{
			"client_username":  clientUsername,
			"partner_username": partnerUsername,
			"limit":            limit,
			"offset":           offset,
		},
	)
	if err != nil {
		log.Println("DMChatModel.go: ChatHistory", err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	messages, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "chat_history")

	return messages, nil
}

func AckMessageDelivered(ctx context.Context, clientUsername, partnerUsername, msgId string, deliveredAt time.Time) error {
	_, err := db.Query(
		ctx,
		`
		MATCH (clientChat:DMChat{ owner_username: $client_username, partner_username: $partner_username }),
      (clientChat)<-[:IN_DM_CHAT]-(message:DMMessage{ id: $message_id, delivery_status: "sent" })<-[:RECEIVES_MESSAGE]-()
    SET message.delivery_status = "delivered", message.delivered_at = $delivered_at, clientChat.unread_messages_count = coalesce(clientChat.unread_messages_count, 0) + 1

		`,
		map[string]any{
			"client_username":  clientUsername,
			"partner_username": partnerUsername,
			"message_id":       msgId,
			"delivered_at":     deliveredAt,
		},
	)
	if err != nil {
		log.Println("DMChatModel.go: AckMessageDelivered", err)
		return fiber.ErrInternalServerError
	}

	return nil
}

func AckMessageRead(ctx context.Context, clientUsername, partnerUsername, msgId string, readAt time.Time) error {
	_, err := db.Query(
		ctx,
		`
		MATCH (clientChat:DMChat{ owner_username: $client_username, partner_username: $partner_username }),
      (clientChat)<-[:IN_DM_CHAT]-(message:DMMessage{ id: $message_id } WHERE message.delivery_status IN ["sent", "delivered"])<-[:RECEIVES_MESSAGE]-()

    WITH clientChat, message, CASE coalesce(clientChat.unread_messages_count, 0) WHEN <> 0 THEN clientChat.unread_messages_count - 1 ELSE 0 END AS unread_messages_count
    SET message.delivery_status = "read", message.read_at = $read_at, clientChat.unread_messages_count = unread_messages_count

		`,
		map[string]any{
			"client_username":  clientUsername,
			"partner_username": partnerUsername,
			"message_id":       msgId,
			"read_at":          readAt,
		},
	)
	if err != nil {
		log.Println("DMChatModel.go: AckMessageRead", err)
		return fiber.ErrInternalServerError
	}

	return nil
}
